package personalDB

const (
	//First 2 bits
	LEAF   = 2
	BRANCH = 1
	//HEADER is 4 bits long. first 2 are the type and the second 2 are the number of keys the node has
	HEADER       = 4
	PAGE_SIZE    = 4096
	MAX_KEY_SIZE = 1000
	MAX_VAL_SIZE = 3000
)
