package trollhash

import (
	"io"
)

type Trollhash func(byte) uint64

func New(size int) Trollhash {
	lshift := (3 * size) % 64
	rshift := 64 - lshift
	history := make([]uint64, size)
	slot := 0
	troll := uint64(0)
	return func(key byte) uint64 {
		remove := history[slot]
		tool := seedlings[key]
		history[slot] = tool
		troll = (troll << 3) | (troll >> 61) ^ tool
		troll ^= (remove << lshift) | (remove >> rshift)
		slot += 1
		if slot == size {
			slot = 0
		}
		return troll
	}
}

func Hash(needle []byte) (result uint64) {
	hasher := New(len(needle))
	for _, add := range needle {
		result = hasher(add)
	}
	return result
}

type WriteLocator interface {
	io.Writer
	Locations() []int64
}

type writer struct {
	delegate io.Writer
	seek     Seeker
	found    []int64
}

func (it *writer) Write(payload []byte) (int, error) {
	for _, value := range payload {
		ok, head := it.seek(value)
		if ok {
			it.found = append(it.found, head)
		}
	}
	return it.delegate.Write(payload)
}

func (it *writer) Locations() []int64 {
	return it.found
}

func LocateWriter(delegate io.Writer, needle string) WriteLocator {
	result := &writer{
		delegate: delegate,
		seek:     Find(needle),
		found:    make([]int64, 0, 20),
	}
	return result
}

type Seeker func(byte) (bool, int64)

func makeSeeker(verify []byte) Seeker {
	goal := Hash(verify)
	limit := int64(len(verify))
	cursor, window := int64(-1), make([]byte, limit)
	hasher := New(int(limit))
	slot, size := 0, int(limit)
	return func(add byte) (bool, int64) {
		cursor += 1
		window[slot] = add
		slot += 1
		if slot == size {
			slot = 0
		}
		if hasher(add) != goal {
			return false, -1
		}
		for at, value := range verify {
			if value != window[(cursor+1+int64(at))%limit] {
				return false, -1
			}
		}
		return true, cursor - limit + 1
	}
}

func Find(needle string) Seeker {
	return makeSeeker([]byte(needle))
}

func Seedlings() map[uint64]int {
	result := make(map[uint64]int)
	for at, value := range seedlings {
		result[value] = at
	}
	return result
}

var seedlings = [256]uint64{
	0x602b8129934c868b, 0x6ab8983d8c8ba01c, 0x54d04ec8dee35ff5, 0xfdc3567e6f901fd5,
	0x2562b4b5b0cc366c, 0xab60180bab46c43f, 0x5a54553f73bd3e1b, 0xaf194e0fe6c4bf87,
	0x25c98c0d6b63a370, 0x8c81fae8b37fff21, 0x53ead3efb8b5b25d, 0xcce5c646782cd9cf,
	0xfdf778b8ae365720, 0x4d977df169207cca, 0x156a5b75bc7798b5, 0x46c7f1335c2ed747,
	0x8310228569b88651, 0x35809eeb9ca50863, 0x441ee622f898ad0a, 0xc5bb8b2cf3932d6b,
	0xf7c606cabd736bc6, 0x28b4e86876eeedb9, 0xb49d9c3599dd647e, 0xda85b43fa53edcf2,
	0x392ee67addd0d02f, 0x73a31ef6d1b95fba, 0x5e169bbb3d28951b, 0x6969557639c6a9e8,
	0xb03c20f52a6f3fd4, 0x27ffff7cf607addd, 0x335c0951697ce069, 0xf40a8371a53c31dc,
	0x13a765069f736cc6, 0x942216377a8d67e3, 0xed0c5da61b167d3a, 0xaa9d5e228b5567a8,
	0x4cd448389d866d4f, 0x93d6a2e30f18e84f, 0xf8be8616379c8db1, 0x95c7c233cc922e36,
	0x8021cb619a850787, 0x7f6d5edea66f25a9, 0x0e6c9d1e6c9646f9, 0x02b9a1ab0b82bb32,
	0x8a4344782e76446f, 0xa93d1c7ff54f35cd, 0x303e4f8594da3e66, 0x8d034c3bcc43340e,
	0x70977337f51155a0, 0x750701470ef3de59, 0x3c57e01aeb3a9e59, 0xd583288188c6289a,
	0x656d0c50ea6bb54d, 0x76bbc1e73deb85ef, 0x40db7a12e5c065a3, 0xce585e8b46166d1d,
	0x284ee3e3fe8aa20b, 0x3f315e2464e9d196, 0x6f5b08ad33872bd9, 0x00b1405c606adb9e,
	0xbf20e1957769961a, 0x297ba8e2d1903af3, 0x6f903a275e60451e, 0x87a83971fb761024,
	0x9348cb5ff1383281, 0x2f203fec99682f8d, 0x1443e52adddcf08e, 0xef09ff3b737a55fc,
	0xe92cd8b27e851a79, 0x6e1e59a28c13f09a, 0x24a7c49a9bade515, 0xff3045ffc77d2b24,
	0xf16f51fa093ccdc6, 0xdf04734eec39671e, 0x76b6f9adb5c2d094, 0x3904a429219a48ef,
	0x9d0244a46fd87f84, 0x53177c2f2b3465d9, 0xf27b02832137d20b, 0x72a45d5c27ef2bde,
	0x8c8307bfc4117674, 0x69ca61ca73e8113d, 0x244eb5285055a241, 0xa57fa8ef3f85efc8,
	0x607d3dab80e04aa0, 0x7e47163b689e8c81, 0xcddc93876c73a1a8, 0x94d4635b1fcaa2e3,
	0xf1283c1bbc591b52, 0x54098cf2b11c8d68, 0x5b181f79ddc50186, 0x77c57d4e824a4636,
	0x39f888233f81fc4e, 0x5eb7eb313d175801, 0xad4b57e527f01949, 0xb91b4230e16f2edc,
	0x92a0676324bf9721, 0x384bf9513d1fb244, 0x40cf2ef187ef03cf, 0x6f12ddbbd8383773,
	0xa3110f4ead8a066a, 0x1b6ab1431567bb2d, 0x73a397be57c72c8a, 0x4c9c445d1db0c18c,
	0xc6cb2c15ed2b1faa, 0x83a988ce9c10d893, 0xaf4fe6805be23828, 0x776a74a4dec3cd7e,
	0xd23d9949d80c389b, 0xdbd6399f812a0c56, 0xc5bb1c90121ca7ce, 0x332cda9b9ff5afb9,
	0x5a6239acaafceabd, 0x0258f2053a726194, 0x142f72cfc81201cf, 0x85af1c1c2b3d0425,
	0x620568ce9f81d404, 0x28cc4d813b157eb7, 0x36637f0bc3f48ed8, 0x563c25e210789612,
	0x0ef4909420333fae, 0x8fd529de3bbc7a70, 0x8989297984fcc92d, 0x482f38b5985919bf,
	0x8b3604417bc20181, 0x8aa0bd889a5496cb, 0x38d69325a3b94aa1, 0x4b3d5b0c247c8316,
	0xc3bb81aa4a2b8f1e, 0x48baa02799ec924b, 0xd287c474ad6a3d0d, 0xf93312bf407b38a3,
	0x1a62a35cacbabca2, 0x266fdbb2033d743e, 0xac80c27e4a760ce1, 0x8995461d77a933c0,
	0x3ad00ba59d07b4f3, 0x969a9e95f9617ae3, 0x1079fe07d879543d, 0x31880d8a6420cb8c,
	0x84eba52cbad9d38a, 0x6477aa7aebf5d8b2, 0x31c5bcc065ae3124, 0x5e1b2121b423868e,
	0x3f208e19c74b0994, 0xc3ed021162e50000, 0x49a71a2f28d1732b, 0xcbe5d2df36846b2b,
	0xf3b59192f546f437, 0x0d826b0d72b2fd15, 0x6eeaa0c0a2ffb7f2, 0x6113e0030b7d5908,
	0x3659c2043e8aee58, 0xf1060b073baf9339, 0xc072daff6bc681f8, 0x884345ae6ef6c538,
	0x202184833fd0bbbf, 0xd266a3bb47fc22f1, 0x8914c38ffeaa392e, 0x73ce11a141170ea1,
	0xc0df84b348e03fbf, 0x6d745be300145ac8, 0x063b321ab6d0fad8, 0xf87bc2d24666b17e,
	0x320cd31b8df1ef33, 0xc1ec122e2fd5bb39, 0x74f5923d5d2b5eba, 0xe85798e5f5cf02c5,
	0x83360efbfe2ffae2, 0x79a0dba643e4b98a, 0x87512888e7e12293, 0xb1fd433d8a37043a,
	0xbad8d58a167d3ca5, 0x99855b2b011c29cc, 0x4fcfde91f1ef652a, 0xec462ce2c5b2a730,
	0xb2b03ac73d5de194, 0x9565e9662275f9aa, 0x1f5a09117923a94d, 0x201b3c78cc80bb5f,
	0x105da69edd31574f, 0x9444318eb5e5af8d, 0xb4bc0f4f23295ef7, 0x86988e0c3546863b,
	0x1827e710952a0df2, 0x84af4a83f577e63e, 0x477bee7db82f88d1, 0xc7808fa972ea660e,
	0x0904fe12acff3f63, 0xb5baf8db3fb767f1, 0xe675059b5b603db1, 0xcf34d51dbddd9733,
	0xf7dc0a20ebc5f184, 0xb49588039d34ee77, 0x3015fa3d8f3145f9, 0x26afeef62e3a9a29,
	0x174257d586d9a3f2, 0x9aa8876a6a5eafd5, 0xde5150df0c3eef53, 0x3ba44df0da99cb4a,
	0xcf8668774135287e, 0xe8abe6669fb8aa4d, 0xe8a72dcf45e862fa, 0xcf11e3d463f05295,
	0xdd8862d9877b910b, 0x1d4cc1684ee326fc, 0x98e7907b726a7b53, 0xad845110e13a3c48,
	0x37de32a13c2d959a, 0xecfbec7d1a2c4339, 0xb2334e53257814a9, 0xe287b65ac7e079d8,
	0x2f44ae0cecfd15ff, 0xf71e440c162d323f, 0x4ebc72f70a4438c4, 0x9972e1d375126584,
	0xdf43537388eecfb8, 0x93f1597d9f115c0b, 0x3e7c5ab2ce23d493, 0x01bcd6f5d1b559ae,
	0xae3a89d1c2ab3c97, 0x224c3ad9e70defa3, 0x8cc7be5774b6a234, 0x07c5a14d71baff0e,
	0xef36700d8e824543, 0x1a730b83ccac66bd, 0x2179a3d23e72bec5, 0xbdd5c3445c00db14,
	0xc9f22df506ca595d, 0x70c58e3b4b014d74, 0x938755c1c7e11634, 0x2ce1a74793a461bd,
	0xbf1f4a2f14f594ef, 0x95ec2bce4bf8481a, 0xe9d8809a5065a5b4, 0x6cec014177d56fbf,
	0xb68c5f723764db37, 0xa2f5f21314b3935f, 0x3d3289499254d863, 0x85a5b5667439987d,
	0x31710194f888f426, 0x61e713b19996c044, 0x922938c7acbf5a69, 0xb7d3bcf2f449d3b9,
	0x24401236b825b103, 0xb6067fe7627e6b5a, 0x74732e6e3fc80d22, 0xa172f36ec342041e,
	0xd1465ccc18cf6df3, 0xf4fb5eb5ccf459a2, 0x6b22cce719a2c185, 0xcf41f79318890218,
	0x1d65d8c00d13b16c, 0xf1d1fd3f8c5c3c81, 0xabf782ddcdc96476, 0x62aed2ded00413ca,
}
