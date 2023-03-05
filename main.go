package main

import (
	"encoding/binary"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"
)

type Mode int

const (
	ENCODE_MODE Mode = iota
	DECODE_MODE Mode = iota
)

type Config struct {
	filename string
	mode     Mode
}

type TreeNode struct {
	freq        int
	value       byte
	left, right *TreeNode
}

func left_pad(_s fmt.Stringer) (res string) {
	s := _s.String()
	for _, line := range strings.Split(s, "\n") {
		res = fmt.Sprintf("%s\n  %s", res, line)
	}
	return
}

func (self TreeNode) String() string {
	children := ""
	if self.left != nil {
		children = fmt.Sprintf("%s%s", left_pad(self.left), left_pad(self.right))
	}
	return fmt.Sprintf("%c(freq %d)%s", self.value, self.freq, children)
}

type BitSeq struct {
	length int
	data   []byte
}

func (self BitSeq) add_bit(bit bool) (res BitSeq) {
	res.data = append([]byte{}, self.data...)
	if (self.length % 8) == 0 {
		res.data = append(res.data, 0)
	}

	if bit {
		res.data[len(res.data)-1] |= (1 << (self.length % 8))
	}

	res.length = self.length + 1
	return
}

func (self *BitSeq) concat(other BitSeq) (res BitSeq) {
	for i := 0; i < other.length; i++ {
		*self = self.add_bit((other.data[i/8] & (1 << (i % 8))) != 0)
	}
	return
}

func usage() {
	fmt.Println("Usage: gopress <input-file> (-d/-e)")
	fmt.Println("-e - encode mode")
	fmt.Println("-d - decode mode")
	os.Exit(1)
}

func contains[T comparable](elems []T, v T) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

func parse_args() (config Config) {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Input file is missing.")
		usage()
	}

	config.filename = args[0]

	is_encode := contains(args, "-e")
	is_decode := contains(args, "-d")
	if !((is_encode || is_decode) && !(is_encode && is_decode)) { // NOT XOR
		fmt.Println("Can't provide -d and -e at the same time.")
		usage()
	}
	if is_encode {
		config.mode = ENCODE_MODE
	} else {
		config.mode = DECODE_MODE
	}

	return
}

func get_freqs(input []byte) (res map[byte]int) {
	res = make(map[byte]int)
	for _, char := range input {
		prev, ok := res[char]
		if ok {
			res[char] = prev + 1
		} else {
			res[char] = 1
		}
	}
	return
}

func make_queue(freqs map[byte]int) (res []TreeNode) {
	for k, v := range freqs {
		res = append(res, TreeNode{freq: v, value: k})
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].freq > res[j].freq
	})
	return
}

func put_node(queue *[]TreeNode, node TreeNode) {
	inserted := false

	for i := range *queue {
		curr := (*queue)[len(*queue)-1-i]
		if curr.freq > node.freq {
			inserted = true
			*queue = append((*queue)[:len(*queue)-i], (*queue)[len(*queue)-i-1:]...)
			(*queue)[len(*queue)-i-1] = node
			break
		}
	}

	if !inserted {
		*queue = append([]TreeNode{node}, *queue...)
	}
}

func make_tree(queue []TreeNode) TreeNode {
	for len(queue) > 1 {
		left, right := queue[len(queue)-1], queue[len(queue)-2]
		queue = queue[:len(queue)-2]
		put_node(&queue, TreeNode{freq: left.freq + right.freq, left: &left, right: &right})
	}

	return queue[0]
}

func make_map(node TreeNode, path BitSeq, encoding_map *map[byte]BitSeq) {
	if node.left == nil {
		(*encoding_map)[node.value] = path
	} else {
		make_map(*node.left, path.add_bit(false), encoding_map)
		make_map(*node.right, path.add_bit(true), encoding_map)
	}
}

// Tree debug
// func generate_dot(node *TreeNode) string {
// 	if node.left == nil {
// 		return fmt.Sprintf("Node%p[label=%#v]", node, fmt.Sprintf("%#v\n%d", fmt.Sprintf("%c", node.value), node.freq))
// 	} else {
// 		node_repr := fmt.Sprintf("Node%p[label=\"%d\"]", node, node.freq)
// 		connections := fmt.Sprintf("Node%[1]p -> Node%[2]p[label=\"0\"]\nNode%[1]p -> Node%[3]p[label=\"1\"]",
// 			node, node.left, node.right)
// 		return fmt.Sprintf("%s\n%s\n%s\n%s", node_repr, connections, generate_dot(node.left), generate_dot(node.right))
// 	}
// }

func generate_content(bits BitSeq) (res []byte) {
	res = append(res, 0, 0, 0, 0)
	binary.LittleEndian.PutUint32(res, uint32(bits.length))
	res = append(res, bits.data...)
	return
}

func generate_map(m map[byte]BitSeq) (res []byte) {
	for k, v := range m {
		res = append(res, 0, 0, 0, 0)
		binary.LittleEndian.PutUint32(res[len(res)-4:], uint32(v.length))
		res = append(res, v.data...)
		res = append(res, k)
	}
	return
}

func encode(config Config, input []byte) {
	encoding_map := make(map[byte]BitSeq)
	make_map(make_tree(make_queue(get_freqs(input))),
		BitSeq{}, &encoding_map)

	res_content := BitSeq{}
	for _, char := range input {
		res_content.concat(encoding_map[char])
	}
	content := append(generate_map(encoding_map), generate_content(res_content)...)
	err := os.WriteFile(fmt.Sprintf("%s.gprs", config.filename), content, fs.FileMode(os.ModePerm))
	if err != nil {
		fmt.Println("Failed to write to file\n Traceback:")
		panic(err)
	}
	// Tree debug
	// os.WriteFile("graph.dot",
	// 	[]byte(fmt.Sprintf("digraph Tree {\n%s\n}", generate_dot(&tree))),
	// 	fs.FileMode(os.O_RDWR))

	// Map debug
	// for k, v := range encoding_map {
	// 	fmt.Printf("%c: %08b(%d)\n", k, v.data, v.length)
	// }

	// Result debug
	// fmt.Printf("%d %08b\n", res_content.length, res_content.data)
}

func main() {
	config := parse_args()

	input, err := os.ReadFile(config.filename)
	if err != nil {
		panic(err)
	}

	if config.mode == ENCODE_MODE {
		encode(config, input)
	} else {
		panic("UNIMPLEMENTED")
	}
}
