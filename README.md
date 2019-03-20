# cfmetadata-plugin
A CF cli plugin to view and modify metadata

## Installation
1. git clone the repo to your desktop
1. In the repo, run `go build` to compile a binary
1. run `cf install-plugin <path-to-binary>`

## Known Issue
Resolved in CLI [v6.43.0](https://github.com/cloudfoundry/cli/releases/tag/v6.43.0).
~~The CF cli has a [bug](https://github.com/cloudfoundry/cli/issues/1108) that causes the user token to periodically expire. This will manifest as not found errors for
resources that exist. To resolve run a normal cli command and then rerun the command from this plugin.~~

## Supported Resources
- Organizations
- Spaces
- Apps

## Usage

### View all metadata 
```
cf metadata app my-app
```
```
cf metadata space my-space
```
```
cf metadata organization my-org
```

### View labels
```
cf labels app my-app
```
```
cf labels space my-space
```
```
cf labels organization my-org
```

### View annotations
```
cf annotations app my-app
```
```
cf annotations space my-space
```
```
cf annotations organization my-org
```

### Manage labels

- Add `environment` label, modify `stable` label,  remove `beta` label

```
cf labels my-app environment=production stable=true beta-
```

### Manage annotations

same as labels
