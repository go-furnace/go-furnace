# ROADMAP
## v1.0.0-beta

- Implement configuration management.

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
