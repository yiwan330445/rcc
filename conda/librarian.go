package conda

type Changes struct {
	Name    string
	Dryrun  bool
	Pip     bool
	Channel bool
	Add     []string
	Remove  []string
}

func UpdateEnvironment(filename string, changes *Changes) (string, error) {
	environment := SummonEnvironment(filename)
	if changes.Channel {
		updateChannels(environment, changes)
	} else {
		err := updatePackages(environment, changes)
		if err != nil {
			return "", err
		}
	}
	if len(changes.Name) > 0 {
		environment.Name = changes.Name
	}
	if changes.Dryrun {
		return environment.AsYaml()
	}
	err := environment.SaveAs(filename)
	if err != nil {
		return "", err
	}
	return environment.AsYaml()
}

func Index(search string, members []string) int {
	for at, member := range members {
		if member == search {
			return at
		}
	}
	return -1
}

func updateChannels(environment *Environment, changes *Changes) {
	result := make([]string, 0, len(changes.Add)+len(environment.Channels))
	for _, current := range environment.Channels {
		if Index(current, changes.Remove) > -1 {
			continue
		}
		result = append(result, current)
	}
	for _, here := range changes.Add {
		if Index(here, result) > -1 {
			continue
		}
		result = append(result, here)
	}
	environment.Channels = result
}

func updatePackages(environment *Environment, changes *Changes) error {
	adds := asDependencies(changes.Add)
	removes := asDependencies(changes.Remove)
	if changes.Pip {
		result, err := composePackages(environment.Pip, adds, removes)
		if err != nil {
			return err
		}
		environment.Pip = result
	} else {
		result, err := composePackages(environment.Conda, adds, removes)
		if err != nil {
			return err
		}
		environment.Conda = result
	}
	return nil
}

func composePackages(target []*Dependency, add []*Dependency, remove []*Dependency) ([]*Dependency, error) {
	result := make([]*Dependency, 0, len(target)+len(add))
	for _, current := range target {
		if current.Index(remove) > -1 {
			continue
		}
		result = append(result, current)
	}
	for _, current := range add {
		found := current.Index(result)
		if found < 0 {
			result = append(result, current)
			continue
		}
		selected, err := current.ChooseSpecific(result[found])
		if err != nil {
			return nil, err
		}
		result[found] = selected
	}
	return result, nil
}

func asDependencies(labels []string) []*Dependency {
	result := make([]*Dependency, 0, len(labels))
	for _, label := range labels {
		result = append(result, AsDependency(label))
	}
	return result
}
