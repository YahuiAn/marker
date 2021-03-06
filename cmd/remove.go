/*
Copyright © 2022 yahuian <yahuian@126.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/yahuian/marker/config"
	"github.com/yahuian/marker/pkg/tree"
)

var (
	yes = false
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove useless images",
	Long: `Remove relative path images that exist on your local computer but are not referenced.
There can be multiple files and directories (support nested) in the root path.
`,
	RunE: runRemove,
	Example: `  marker remove
  marker remove -y
`,
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().BoolVarP(&yes, "yes", "y", false, "When yes is false only show useless images.")
}

func runRemove(cmd *cobra.Command, args []string) error {
	fsys := os.DirFS(root)
	t, err := tree.NewTree(fsys, func(d fs.DirEntry) bool {
		return config.SkipFiles(d)
	})
	if err != nil {
		return fmt.Errorf("new tree err: %w", err)
	}
	images, err := t.GetUselessImages(fsys, config.Val.ImageTypes)
	if err != nil {
		return fmt.Errorf("get useless images err: %w", err)
	}

	if len(images) == 0 {
		fmt.Println("Well done, your images are all used.")
		return nil
	}

	if !yes {
		fmt.Println("These images are useless, you can remove them with --yes flag.")
	}
	for _, v := range images {
		if !yes {
			fmt.Println(v)
			continue
		}

		p := path.Join(root, v)
		if err := os.Remove(p); err != nil {
			return fmt.Errorf("remove %s err: %w", p, err)
		}

		fmt.Printf("[removed] %s\n", p)
	}

	return nil
}
