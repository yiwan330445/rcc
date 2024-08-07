import os, sys, subprocess
from os import chdir
from subprocess import run

task = '-T' if len(sys.argv) < 2 else sys.argv[1]
os.chdir('..')
exit(subprocess.run(('rake', task)).returncode)
