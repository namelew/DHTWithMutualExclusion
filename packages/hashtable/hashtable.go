package hashtable

import (
	"errors"

	"github.com/tidwall/btree"
	"golang.org/x/exp/constraints"
)

type Key[K constraints.Ordered] interface {
	Value() K
	Index() int
}

type HashTable[I constraints.Ordered, K comparable] interface {
	Hash(id Key[I]) int
	setArguments(atributes Common)
	Insert(id Key[I], data K) error
	Delete(id Key[I]) error
	Search(id Key[I]) (K, error)
}

type Common struct {
	End  int
	Size int
}

type Linked[I constraints.Ordered, K comparable] struct {
	size  int
	slots []btree.Map[I, K]
}

type Open[I constraints.Ordered, K comparable] struct {
	end     int
	size    int
	indices []I
	slots   []K
}

func New[I constraints.Ordered, K comparable](hashTable HashTable[I, K], parameters Common) HashTable[I, K] {
	hashTable.setArguments(parameters)
	return hashTable
}

func (t *Linked[I, K]) Hash(id Key[I]) int {
	return id.Index() % t.size
}

func (t *Linked[I, K]) setArguments(atributes Common) {
	t.size = atributes.Size
	t.slots = make([]btree.Map[I, K], t.size)
}

func (t *Linked[I, K]) Insert(id Key[I], data K) error {
	t.slots[t.Hash(id)].Set(id.Value(), data)
	return nil
}

func (t *Linked[I, K]) Delete(id Key[I]) error {
	_, found := t.slots[t.Hash(id)].Delete(id.Value())

	if !found {
		return errors.New("delete action failed: unable to find register")
	}

	return nil
}

func (t *Linked[I, K]) Search(id Key[I]) (K, error) {
	result, found := t.slots[t.Hash(id)].Get(id.Value())

	if !found {
		return result, errors.New("query action failed: unable to find register")
	}

	return result, nil
}

func (t *Open[I, K]) Hash(id Key[I]) int {
	return id.Index() % t.size
}

func (t *Open[I, K]) setArguments(atributes Common) {
	t.size = atributes.Size
	t.end = atributes.End
	t.slots = make([]K, t.size)
	t.indices = make([]I, t.size)
}

func (t *Open[I, K]) Insert(id Key[I], data K) error {
	var empty K
	slot := t.Hash(id)

	if t.slots[slot] != empty {
		if t.indices[slot] == id.Value() {
			return errors.New("unable to insert data: duplicated key")
		}
		for i := slot + 1; i < t.end; i++ {
			if t.slots[i] == empty {
				t.slots[i] = data
				t.indices[i] = id.Value()
				return nil
			}
		}
		return errors.New("unable to insert data: no free slot")
	}

	t.slots[slot] = data
	t.indices[slot] = id.Value()
	return nil
}

func (t *Open[I, K]) Delete(id Key[I]) error {
	var empty K
	var emptyID I
	slot := t.Hash(id)

	if t.indices[slot] == id.Value() {
		t.slots[slot] = empty
		t.indices[slot] = emptyID
		return nil
	} else {
		for i := slot + 1; i < t.end; i++ {
			if t.indices[i] == id.Value() {
				t.slots[i] = empty
				t.indices[i] = emptyID
				return nil
			}
		}
	}

	return errors.New("unable to find element")
}

func (t *Open[I, K]) Search(id Key[I]) (K, error) {
	var empty K
	slot := t.Hash(id)

	if t.indices[slot] == id.Value() {
		return t.slots[slot], nil
	} else {
		for i := slot + 1; i < t.size; i++ {
			if t.indices[i] == id.Value() {
				return t.slots[i], nil
			}
		}
	}

	return empty, errors.New("unable to find element")
}
