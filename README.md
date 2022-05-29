# hist-sanitizer

This tool which builds the changelog for the Swift repository between two given tags, eliminating Merge commits and replacing them with the title of the Pull Request which created the Merge commit.

#### Exemple:

```
* Merge pull request #PR_ID from 'feature/another'
* Title of commit 4
* Merge pull request #PR_ID from 'feature/branch'
* Title of commit 2
* Title of commit 1
```

is transformed into:

```
* Title of the Pull Request
* Title of commit 4
* Title of the Pull Request
* Title of commit 2
* Title of commit 1
```

## How to use it:
Just call it from the cli replacing the tags:
``` bash
$ ./hist-sanitizer <GIT-TAG-BASE> <GIT-TAG-HEAD>
```
> Ex:. ./hist-sanitizer swift-5.5.3-RELEASE swift-5.5.1-RELEASE

Or for authenticated calls:
``` bash
$ GITHUB_API_TOKEN=<your-token> ./hist-sanitizer <GIT-TAG-BASE> <GIT-TAG-HEAD>
```

## Development mode:
Using docker with live reloading

On `.air.toml` file change the property `full_bin` passing the tags you want to see on live reload
Then:
``` bash
$ docker compose up
```
> For using authenticated user Create .env file with `GITHUB_API_TOKEN` property
