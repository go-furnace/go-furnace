# ROADMAP

## v1.1.0-beta

- `update` command which updates the stack with new information
- `rollout` command which rotates an outdated instance out of the cloudformation stack and adds a new one
- Poller - currently on errors the application always dies. A poller would be nice which retries certain requests a configured
  number of times.
- Timeout with context? Currently most of the actions don't time out, or use a default timeout from aws. Thus they either run
  foever or 15 minutes tops which is still too much.

## v1.0.0-beta

- Implement configuration management support for chef, puppet, ansible.

## v0.9.0-beta

- Add git revision configuration.
- Add deploying from S3 bucket.
    - For this the cf config needs access to the bucket.
- Add control over the current environment to the plugin system.

## v0.0.1

- Add error test cases for the calls which return errors.
    - For example describeStacks returns an error if the stack is non-existent.
- Add push command which deals with pushing a version of the application to an
existing stack.
