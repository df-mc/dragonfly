# Contributing to Dragonfly

First of all, thank you for your interest in contributing to Dragonfly. :+1:

The following is a set of guidelines for contributing to Dragonfly. These are guidelines and in
general it is recommended to stick to them when contributing, but you should use your best
judgement. Feel free to propose changes to this document in our Discord.

### Pull Requests
In general, it is recommended to discuss any changes you would like to make in our Discord to
before making changes, unless the change-set is otherwise small or limited to a specific part of
the code base.

When reviewing pull requests, we aim to reach the following goals with the code proposed:
* Maintain the quality of the source code.
* Stick to the standard formatting of Go (go fmt).
* Provide a well-documented, simple and clean codebase.

To make sure your pull request satisfies those points, we recommend you do the following before
opening a pull request:
* Run `go fmt` on any files you have changed to ensure the formatting is correct. Some IDEs have
  integration with this tool to run it automatically when committing. GoLand has a box that may be
  checked in the bottom-right corner when creating a commit.
* Make sure to provide documentation for symbols where adequate. We generally follow the following
  conventions for documentation in pull requests:
  - Exported symbols (TypeName, FunctionName) should always have documentation over them, but if
    the function exists merely to satisfy an interface, the documentation may read 
    `// FunctionName ...`.
  - Unexported symbols (typeName, functionName) _should_ have documentation, but doing so is not
    mandatory if the function is very simple and needs no clarification.
* Make sure to use British English and proper punctuation throughout symbol names, variables and
  documentation.
* Where possible, try to expose as few exported symbols (functions, types) as possible, unless 
  strictly necessary. This makes it easier for us to change code in the future and ensures that 
  users cannot use functions not suitable for the API.
* We strive to have only completely functional features in the codebase. While we recognise that
  it is not always possible to provide full functionality for a feature in a single pull request,
  you should attempt to do so to the extent that you can. Specific smaller features part of the
  pull request that cannot be implemented yet should be marked with a `// TODO: ...` comment so
  that we can implement these once the required functionality is present.
* When you open a PR, we assume you have tested your code and made sure it is working as intended.
  As a general recommendation, you should enable the Minecraft Content Log in the Profile settings
  so that it becomes obvious when invalid data is sent to the client.
  
If you run into a problem or otherwise need help with your pull request, please feel free to reach
out to us on Discord, so we can work towards a complete pull request together.
