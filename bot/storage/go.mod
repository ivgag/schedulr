module github.com/ivgag/schedulr/storage

go 1.23.1

require github.com/ivgag/schedulr/model v0.0.0

replace (
	github.com/ivgag/schedulr/model => ../model
	github.com/ivgag/schedulr/utils => ../utils
)
