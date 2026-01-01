package codegen

import "fmt"

// We model the heap as:
// map[int]map[string]interface{}
//
// Why?
// - Object identity = int (Boogie-style references)
// - Fields are named (string)
// - Values may be int or bool
// - Type safety is enforced by the EBS checker, not at runtime

// Heap represents a simple object heap:
//
//	object id -> field name -> value
var heap = make(map[int]map[string]interface{})

// heapRead reads a field from an object.
// Panics if the object or field is missing.
func heapRead(obj interface{}, field string) interface{} {
	id, ok := obj.(int)
	if !ok {
		panic("heapRead: object is not an int reference")
	}

	fields, ok := heap[id]
	if !ok {
		panic(fmt.Sprintf("heapRead: unknown object %d", id))
	}

	val, ok := fields[field]
	if !ok {
		panic(fmt.Sprintf("heapRead: unknown field %s on object %d", field, id))
	}

	return val
}

// heapWrite writes a field on an object.
// Allocates the object if it does not exist.
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
