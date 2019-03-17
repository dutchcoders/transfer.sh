package a

type Ifc interface {
	A(string) string
	B(int) int
	C(chan int) chan int
	D(interface{})
	E(map[string]interface{})
	F([]float64)
}
