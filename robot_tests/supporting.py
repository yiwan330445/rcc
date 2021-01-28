import json
import subprocess

def run_and_return_code_output_error(command):
    task = subprocess.Popen(command, shell=True, stderr=subprocess.PIPE, stdout=subprocess.PIPE)
    out, err = task.communicate()
    return task.returncode, out.decode(), err.decode()

def parse_json(content):
    parsed = json.loads(content)
    assert isinstance(parsed, (list, dict)), f'Expecting list or dict; got {parsed!r}'
    return parsed
