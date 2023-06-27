import json
import glob

configs = {}
for filename in glob.glob('examples/*'):
    if 'local' in filename:
        continue
    with open(filename) as f:
        configs[filename.split('/')[1]] = f.read()

with open('static/nginx_configs.json', 'w') as f:
    f.write(json.dumps(configs))
