# dependencyinjection

A simple dependency injection library using Golang generics

## Usage

The library has two main concepts, `Container` and `Provider`. 

A `Container` is the object that stores all the dependencies, a global container is provided `dependencyinjection.Global`. 

A `Provider` provides a dependency, there are currently three implemented Provided types, `Singleton`, `Factory` and `Value`:
- *Singleton* - Constructor will only be called once, result is used for all subsequent dependency resolutions.
- *Factory* - Constructor is run _every_ time the dependency is resolved.
- *Value* - The static provided value is returned when the dependency is resolved.

Custom Provider types can be used also, by implementing the `Provider` interface. This could allow for more complex strategies.

### Example

In this simple example a type `Config` is added to the global container. This an then later be resolved by objects that require it using `dependencyinjection.Get[Config](ctx, dependencyinjection.Global)`.

```go
func init() {
    // Register Config into the dependency injection container, in this example it loads the config from the filesystem.
    dependencyinjection.MustRegisterSingleton(dependencyinjection.Global, func(ctx context.Context, container *dependencyinjection.Container) (Config, error) {
        // Get filesystem, this allows different implementations to be provided. In reality for config files the real
        // filesystem will be the one used 99% of the time, but this illustrates how to get dependencies.
        //
        // An example of how the fs.FS object could have been added to the container is as follows:
        //  dependencyinjection.MustRegisterValue(dependencyinjection.Global, os.DirFS("/some/path"))
        //
        filesystem, err := dependencyinjection.Get[fs.FS](ctx, container)
        if err != nil {
            return Config{}, err
        }

        // Load the config
        file, err := filesystem.Open("config.json")
        if err != nil {
            return Config{}, err
        }

        // Ensure file gets closed
        defer file.Close()

        // Parse the config from a file
        var config Config
        if err := json.NewDecoder(file).Decode(&config); err != nil {
            return Config{}, err
        }

        // Return loaded config
        return config, nil
    })
}
```
