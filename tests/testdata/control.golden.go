package main

// ========================
// Runtime Heap
// ========================

var heap = make(map[int]map[string]interface{})

func heapRead(obj interface{}, field string) interface{} {
	id, ok := obj.(int)
	if !ok {
		panic("heapRead: object is not an int reference")
	}

	fields, ok := heap[id]
	if !ok {
		panic("heapRead: unknown object")
	}

	val, ok := fields[field]
	if !ok {
		panic("heapRead: unknown field")
	}

	return val
}

func heapWrite(obj interface{}, field string, value interface{}) {
	id, ok := obj.(int)
	if !ok {
		panic("heapWrite: object is not an int reference")
	}

	fields, ok := heap[id]
	if !ok {
		fields = make(map[string]interface{})
		heap[id] = fields
	}

	fields[field] = value
}

func abs(x int) (y int) {
	if x < 0 {
		y = -x
	} else {
		y = x
	}
	return
}
