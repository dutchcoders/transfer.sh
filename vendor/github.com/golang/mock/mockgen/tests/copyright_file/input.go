//go:generate mockgen -package empty_interface -destination mock.go -source input.go -copyright_file=LICENSE

package empty_interface

type Empty interface{}
