# Janky Shimmer

`border-shimmer` allows users to create a subtle effect where the borders around
their windows (managed by the github.com/FelixKratz/JankyBorders software) can
contain an animated transition among a sequence of user-configurable colors.

## Demo

Here's what my custom configuration looks like:

https://github.com/user-attachments/assets/ba665fa6-488c-4646-bd1e-02591d27de11

## Build and Run Locally

To build the binary from source:

``` shell
$> cd janky-shimmer
$> go build -o border-shimmer cmd/border-shimmer/border-shimmer.go
```

To run the executable with the default coniguration:

``` shell
$> ./border-shimmer &
```
