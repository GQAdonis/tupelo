## Dependencies

* cargo installed (rust CLI tool)
* [dep](https://github.com/golang/dep) installed (go dependency tool).

## Contributing

In order to run the code/tests you need to build the libindy-crypto rust library:

```
git submodule update
cd indy-crypto/libindy-crypto
cargo build --release
```

And grab the go dependencies:

```
dep ensure
```
