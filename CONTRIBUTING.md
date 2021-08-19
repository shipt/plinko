# Contributing to Plinko
We want to make contributing to this project simple and convenient.

## Code of Conduct
We have adopted adopted a Code of Conduct that we expect project participants to
adhere to. Please [read the full text](https://code.fb.com/codeofconduct/) so
that you can understand what actions will and will not be tolerated.

## Our Development Process
Shipt's internal repository remains the source of truth. It is
automatically synchronized with GitHub. Contributions can be made through
regular GitHub pull requests.

## Pull Requests
We actively welcome your pull requests. If you are planning on doing a larger
chunk of work or want to change an external facing API, make sure to file an
issue first to get feedback on your idea.

1. Fork the repo and create your branch from `main`.
2. If you've added code that should be tested, add tests.
3. Ensure the test suite passes and your code lints.
4. Consider quashing your commits (`git rebase -i`). One intent alongside one
   commit makes it clearer for people to review and easier to understand your
   intention.

## Copyright Notice for files
Copy and paste this to the top of your new file(s):

```
/**
 * Copyright (c) Shipt.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
```


## Issues
We use GitHub issues to track public bugs. Please ensure your description is
clear and has sufficient instructions to be able to reproduce the issue.

## Coding Style
* The Plinko coding style is generally based on the
  [Golang Effective Go patterns](https://golang.org/doc/effective_go).
* Match the style you see used in the rest of the project. This includes
  formatting, naming things in code, naming things in documentation.
* Run `golangci-lint run` in the project to ensure compliance rules.

## License
By contributing to Plinko, you agree that your contributions will be licensed
under the [LICENSE.md](./LICENSE.MD) file in the root directory of this source tree.