package tree

import (
	"bufio"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/yahuian/marker/utils"
)

type Tree struct {
	Name  string  `json:"name,omitempty"`
	Sons  []*Tree `json:"sons,omitempty"`
	Dir   bool    `json:"dir,omitempty"`
	Refer int     `json:"refer,omitempty"`

	father *Tree
	sons   map[string]*Tree
}

func NewTree(fsys fs.FS, skip func(d fs.DirEntry) bool) (*Tree, error) {
	root := &Tree{
		Name: ".",
		Dir:  true,
		sons: make(map[string]*Tree),
	}

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.Name() == "." {
			return nil
		}

		// skip files
		if skip != nil && skip(d) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		root.insert(path, d.IsDir())

		return nil
	})
	if err != nil {
		return nil, err
	}

	return root, nil
}

func (root *Tree) insert(path string, dir bool) {
	tmp := root
	for _, v := range strings.Split(path, "/") {
		if _, ok := tmp.sons[v]; !ok {
			node := &Tree{
				Name:   v,
				Dir:    dir,
				sons:   make(map[string]*Tree),
				father: tmp,
			}
			tmp.Sons = append(tmp.Sons, node)
			tmp.sons[v] = node
		}
		tmp = tmp.sons[v]
	}
}

// Search support relative path `..` and `.`
// when path is searched it will return subtree root node otherwise return nil
func (root *Tree) Search(path string) *Tree {
	tmp := root
	for _, v := range strings.Split(path, "/") {
		if v == "." {
			continue
		}

		if v == ".." {
			tmp = tmp.father
			continue
		}

		node, ok := tmp.sons[v]
		if ok {
			tmp = node
		} else {
			return nil
		}
	}
	return tmp
}

func (root *Tree) AbsPath() string {
	tmp := root
	if tmp == nil {
		return ""
	}

	p := tmp.Name
	for tmp.father != nil {
		p = path.Join(tmp.father.Name, p)
		tmp = tmp.father
	}

	return p
}

func (root *Tree) GetUselessImages(fsys fs.FS, imageTypes []string) ([]string, error) {
	if err := root.scanMarkdown(fsys); err != nil {
		return nil, fmt.Errorf("scan markdown files err: %w", err)
	}

	var res []string
	queue := root.Sons
	for len(queue) != 0 {
		node := queue[0]
		if !node.Dir && node.Refer < 1 && in(imageTypes, path.Ext(node.Name)) {
			res = append(res, node.AbsPath())
		}
		queue = append(queue[1:], node.Sons...)
	}

	return res, nil
}

func in(list []string, s string) bool {
	for _, v := range list {
		if s == v {
			return true
		}
	}
	return false
}

// scanMarkdown scan markdown files and count images refer number
func (root *Tree) scanMarkdown(fsys fs.FS) error {
	queue := root.Sons

	for len(queue) != 0 {
		node := queue[0]

		if !node.Dir && path.Ext(node.Name) == ".md" {
			images, err := getImages(fsys, node.AbsPath())
			if err != nil {
				return err
			}

			for _, v := range images {
				t := node.father.Search(v)
				if t != nil {
					t.Refer++
				}
			}
		}

		queue = append(queue[1:], node.Sons...)
	}

	return nil
}

func getImages(fsys fs.FS, path string) ([]string, error) {
	images := make([]string, 0)

	file, err := fsys.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scan := bufio.NewScanner(file)
	for scan.Scan() {
		image := utils.ParseMarkdownImage(scan.Text())
		if image != "" {
			images = append(images, image)
		}
	}

	return images, nil
}