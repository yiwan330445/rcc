import json

def parse_json(content):
    parsed = json.loads(content)
    assert isinstance(parsed, (list, dict)), f'Expecting list or dict; got {parsed!r}'
    return parsed
