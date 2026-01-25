# GO-TEMPLATE

An hyper-opinionated, ultra-minimal Golang repository template.

## Getting started 

### Setup

This repository depends on [asdf]. Once installed, please
install all the [asdf plugins] used in this repository, by running
(or something similar):

```shell
$ cut -d' ' -f1 .tool-versions | \
  while read PLUGIN; do
  	  asdf plugin add $PLUGIN || true
  done
```

Then, run:

```shell
$ asdf install
```

### Toolchain

This codebase uses [task] to define a simple "development toolchain"
for the project. It was installed by [asdf] during the [setup](#setup).

Start by typing:

```shell
$ task <TAB><TAB>
``` 

> [!NOTE]
> Setup [task completion] for the best experience.

## License

[MIT](./LICENSE)


[asdf]: https://asdf-vm.com/
[asdf plugins]: https://asdf-vm.com/manage/plugins.html
[task]: https://taskfile.dev/
[task completion]: https://taskfile.dev/docs/installation#setup-completions