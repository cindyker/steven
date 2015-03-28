package main

import (
	"fmt"
	"strings"
)

var (
	nextBlockID   int
	blocks        [0x10000]Block
	blockSetsByID [0x100]*BlockSet
	missingBlock  = &baseBlock{
		name:        "missing_block", // TODO
		plugin:      "steven",
		cullAgainst: true,
	}
)

// Block is a type of tile in the world. All blocks, excluding the special
// 'missing block', belong to a set.
type Block interface {
	// Is returns whether this block is a member of the passed Set
	Is(s *BlockSet) bool

	Plugin() string
	Name() string

	ModelName() string
	ModelVariant() string

	String() string

	ShouldCullAgainst() bool

	setState(key string, val interface{})
	clone() Block
	toData() int
}

// BlockSet is a collection of Blocks.
type BlockSet struct {
	ID int

	Blocks       []Block
	supportsData bool
}

// base of most (if not all) blocks
type baseBlock struct {
	plugin, name string
	Parent       *BlockSet
	cullAgainst  bool
}

// Is returns whether this block is a member of the passed Set
func (b *baseBlock) Is(s *BlockSet) bool {
	return b.Parent == s
}

func (b *baseBlock) init(name string) {
	// plugin:name format
	if strings.ContainsRune(name, ':') {
		pos := strings.IndexRune(name, ':')
		b.plugin = name[:pos]
		b.name = name[pos+1:]
		return
	}
	b.name = name
	b.plugin = "minecraft"
	b.cullAgainst = true
}

func (b *baseBlock) String() string {
	return fmt.Sprintf("%s:%s", b.plugin, b.name)
}

func (b *baseBlock) Plugin() string {
	return b.plugin
}

func (b *baseBlock) Name() string {
	return b.name
}

func (b *baseBlock) ModelName() string {
	return b.name
}
func (b *baseBlock) ModelVariant() string {
	return "normal"
}

func (b *baseBlock) toData() int {
	return 0
}

func (b *baseBlock) ShouldCullAgainst() bool {
	return b.cullAgainst
}

func (b *baseBlock) setState(key string, val interface{}) {
	panic("base has no state")
}

func (b *baseBlock) clone() Block {
	return &baseBlock{
		plugin: b.plugin,
		name:   b.name,
		Parent: b.Parent,
	}
}

// GetBlockByCombinedID returns the block with the matching combined id.
// The combined id is:
//     block id << 4 | data
func GetBlockByCombinedID(id uint16) Block {
	b := blocks[id]
	if b == nil {
		return missingBlock
	}
	return b
}

func alloc(initial Block) *BlockSet {
	id := nextBlockID
	nextBlockID++
	bs := &BlockSet{
		ID:     id,
		Blocks: []Block{initial},
	}
	blockSetsByID[id] = bs
	return bs
}

func (bs *BlockSet) state(sc stateCollection) *BlockSet {
	old := bs.Blocks
	vals := sc.values()
	bs.Blocks = make([]Block, 0, len(old)*len(vals))
	for _, val := range vals {
		for _, o := range old {
			// allocate a new block
			nb := o.clone()
			// add the new state
			nb.setState(sc.key(), val)
			// now add back to the set
			bs.Blocks = append(bs.Blocks, nb)
		}
	}
	return bs
}

func init() {
	// Flatten the ids
	for _, bs := range blockSetsByID {
		if bs == nil {
			continue
		}
		if !bs.supportsData {
			blocks[bs.ID<<4] = bs.Blocks[0]
			continue
		}
		for _, b := range bs.Blocks {
			data := b.toData()
			if data != -1 {
				blocks[(bs.ID<<4)|data] = b
			}
		}
	}
}