package samplepkg1

func e(c chan int) {
	a := <-c
}

func a(c chan int) {
	a := <-c
}
