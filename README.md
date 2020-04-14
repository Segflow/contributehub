
# Description 
contributehub automatically listen to Github event looking for new Go repositories, pull the package and run various analyzers.

Analyzers can fix the code. If the repo got changed a PR gets created.

Analyzers statically analyses the code by creating the Abstract Syntax Tree (AST) and mutate it.


# Analyzers

## channel direction

Channel direction checks check usage of channel direction and make the approriate changes.

### Examples:

The following two functions
```
func B(ch chan int) {
	ch <- 2
}

func BB(ch chan int) {
	for x := range ch {
		_ = x
	}
}
```

Will be changed to 
```
func B(ch chan<- int) {
	ch <- 2
}

func BB(ch <-chan int) {
	for x := range ch {
		_ = x
	}
}
```
