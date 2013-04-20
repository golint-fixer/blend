package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/mewmew/blend"
	"github.com/mewmew/blend/block"
)

// genParse generates the block parser logic required to parse the provided
// blend file.
//
// The output is stored in "parse.go".
func genParse(b *blend.Blend, dna *block.DNA) (err error) {
	f, err := os.Create("parse.go")
	if err != nil {
		return err
	}
	defer f.Close()

	// Create sorted list of type names.
	structs := make(map[string]bool)
	for _, st := range dna.Structs {
		structs[st.Type] = true
	}
	typeNames := make([]string, len(structs))
	var i int
	for typeName := range structs {
		typeNames[i] = typeName
		i++
	}
	sort.Strings(typeNames)

	// Generate start of struct.go file.
	fmt.Fprintf(f, preStruct, b.Hdr.Ver)

	// Generate block body parsing cases.
	for _, typeName := range typeNames {
		typ := strings.Title(typeName)
		fmt.Fprintf(f, midStruct, typeName, typ, typ, typ)
	}

	// Generate end of struct.go file.
	fmt.Fprint(f, postStruct)

	return nil
}

const preStruct = `// NOTE: generated automatically by blendef for Blender v%d.

package block

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

// ParseBody parses the block body and stores it in blk.Body. It is safe to call
// ParseBody multiple times on the same block.
func (blk *Block) ParseBody(order binary.ByteOrder, dna *DNA) (err error) {
	// Get block body reader.
	r, ok := blk.Body.(io.Reader)
	if !ok {
		// Body has already been parsed.
		return nil
	}

	index := blk.Hdr.SDNAIndex
	if index == 0 {
		// Parse based on block code.
		switch blk.Hdr.Code {
		case CodeDATA:
			blk.Body, err = ioutil.ReadAll(r)
			if err != nil {
				return err
			}
		case CodeDNA1:
			blk.Body, err = ParseDNA(r, order)
			if err != nil {
				return err
			}
		case CodeREND, CodeTEST:
			/// TODO: implement specific block body parsing for REND and TEST.
			blk.Body, err = ioutil.ReadAll(r)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("Block.ParseBody: parsing of %%q not yet implemented.", blk.Hdr.Code)
		}
	} else {
		// Parse based on SDNA index.
		typ := dna.Structs[index].Type
		switch typ {
`

const midStruct = `		case "%s":
			if blk.Hdr.Count > 1 {
				// Parse block body structures.
				bodies := make([]*%s, blk.Hdr.Count)
				for i := range bodies {
					body := new(%s)
					err = binary.Read(r, order, body)
					if err != nil {
						return err
					}
					bodies[i] = body
				}
				blk.Body = bodies
			} else {
				// Parse block body structure.
				body := new(%s)
				err = binary.Read(r, order, body)
				if err != nil {
					return err
				}
				blk.Body = body
			}
			/// ### [ tmp ] ###
			// Verify that all bytes in the block body have been read.
			buf, err := ioutil.ReadAll(r)
			if err != nil {
				return err
			}
			if len(buf) > 0 {
				log.Printf("%%d unread bytes in %%q.", len(buf), typ)
				log.Printf("blk.Hdr: %%#v\n", blk.Hdr)
				log.Println(hex.Dump(buf))
				os.Exit(1)
			}
			/// ### [/ tmp ] ###
`

const postStruct = `		}
	}

	return nil
}
`
