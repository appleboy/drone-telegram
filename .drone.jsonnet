local pipeline = import 'pipeline.libsonnet';
local name = 'drone-telegram';

[
  pipeline.test,
  pipeline.build(name, 'linux', 'amd64'),
  pipeline.build(name, 'linux', 'arm64'),
  pipeline.build(name, 'linux', 'arm'),
  pipeline.release,
  pipeline.notifications(depends_on=[
    'linux-amd64',
    'linux-arm64',
    'linux-arm',
    'release-binary',
  ]),
  pipeline.signature('9a4dcc3659b6f2cb98486e40e4cb0c16d6fc19ad783d3bca13d30c476daf8213'),
]
