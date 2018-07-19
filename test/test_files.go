package test

import (
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/ipfs/go-ipfs-cmdkit/files"
)

const shardingTestDir = "shardTesting"
const shardingTestTree = "testTree"
const shardingTestFile = "testFile"

// Variables related to adding the testing directory generated by tests
var (
	NumShardingDirPrints       = 15
	ShardingDirBalancedRootCID = "QmbfGRPTUd7L1xsAZZ1A3kUFP1zkEZ9kHdb6AGaajBzGGX"
	ShardingDirTrickleRootCID  = "QmcqtKBVCrgZBXksfYzUxmw6S2rkyQhEhckqFBAUBcS1qz"

	// These hashes should match all the blocks produced when adding
	// the files resulting from GetShardingDir*
	// They have been obtained by adding the "shardTesting" folder
	// to go-ipfs (with wrap=true and default parameters). Then doing
	// `refs -r` on the result. It contains the wrapping folder hash.
	ShardingDirCids = [29]string{
		"QmbfGRPTUd7L1xsAZZ1A3kUFP1zkEZ9kHdb6AGaajBzGGX",
		"QmdHXJgxeCFf6qDZqYYmMesV2DbZCVPEdEhj2oVTxP1y7Y",
		"QmSpZcKTgfsxyL7nyjzTNB1gAWmGYC2t8kRPpZSG1ZbTkY",
		"QmSijPKAE61CUs57wWU2M4YxkSaRogQxYRtHoEzP2uRaQt",
		"QmYr6r514Pt8HbsFjwompLrHMyZEYg6aXfsv59Ys8uzLpr",
		"QmfEeHL3iwDE8XjeFq9HDu2B8Dfu8L94y7HUB5sh5vN9TB",
		"QmTz2gUzUNQnH3i818MAJPMLjBfRXZxoZbdNYT1K66LnZN",
		"QmPZLJ3CZYgxH4K1w5jdbAdxJynXn5TCB4kHy7u8uHC3fy",
		"QmUNLLsPACCz1vLxQVkXqqLX5R1X345qqfHbsf67hvA3Nn",
		"QmY6PArrjY66Nb4qEKWF7RUHCToRFyTsrM6cH8D6vJMSnk",
		"QmYXgh47x4gr1iL6YRqAA8RcE3XNWPfB5VJTt9dBfRnRHX",
		"QmXqkKUxgWsgXEUsxDJcs2hUrSrFnPkKyGnGdxpm1cb2me",
		"Qmbne4XHMAiZwoFYdnGrdcW3UBYA7UnFE9WoDwEjG3deZH",
		"Qmdz4kLZUjfGBSvfMxTQpcxjz2aZqupnF9KjKGpAuaZ4nT",
		"QmavW3cdGuSfYMEQiBDfobwVtPEjUnML2Ry1q8w8X3Q8Wj",
		"QmfPHRbeerRWgbu5BzxwK7UhmJGqGvZNxuFoMCUFTuhG3H",
		"QmaYNfhw7L7KWX7LYpwWt1bh6Gq2p7z1tic35PnDRnqyBf",
		"QmWWwH1GKMh6GmFQunjq7CHjr4g4z6Q4xHyDVfuZGX7MyU",
		"QmVpHQGMF5PLsvfgj8bGo9q2YyLRPMvfu1uTb3DgREFtUc",
		"QmUrdAn4Mx4kNioX9juLgwQotwFfxeo5doUNnLJrQynBEN",
		"QmdJ86B7J8mfGq6SjQy8Jz7r5x1cLcXc9M2a7T7NmSMVZx",
		"QmS77cTMdyx8P7rP2Gij6azgYPpjp2J34EVYuhB6mfjrQh",
		"QmbsBsDspFcqi7xJ4xPxcNYnduzQ5UQDw9y6trQWZGoEHq",
		"QmakAXHMeyE6fHHaeqicSKVMM2QyuGbS2g8dgUA7ns8gSY",
		"QmTC6vGbH9ABkpXfrMmYkXbxEqH12jEVGpvGzibGZEDVHK",
		"QmebQW6nfE5cPb85ZUGrSyqbFsVYwfuKsX8Ur3NWwfmnYk",
		"QmSCcsb4mNMz3CXvVjPdc7kxrx4PbitrcRN8ocmyg62oit",
		"QmZ2iUT3W7jh8QNnpWSiMZ1QYgpommCSQFZiPY5VdoCHyv",
		"QmdmUbN9JS3BK3nvcycyzFUBJqXip5zf7bdKbYM3p14e9h",
	}
)

// ShardingTestHelper helps generating files and folders to test adding and
// sharding in IPFS Cluster
type ShardingTestHelper struct {
	randSrc *rand.Rand
}

// NewShardingTestHelper returns a new helper.
func NewShardingTestHelper() *ShardingTestHelper {
	return &ShardingTestHelper{
		randSrc: rand.New(rand.NewSource(1)),
	}
}

func folderExists(t *testing.T, path string) bool {
	if st, err := os.Stat(path); os.IsNotExist(err) {
		return false
	} else if err != nil {
		t.Fatal(err)
	} else if !st.IsDir() {
		t.Fatalf("%s is not a directory", path)
	}
	return true
}

func makeDir(t *testing.T, path string) {
	if !folderExists(t, path) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
	}
}

// see GetTreeMultiReader
func (sth *ShardingTestHelper) makeTestFolder(t *testing.T) {
	makeDir(t, shardingTestDir)
}

// This produces this:
// - shardTesting
//   - testTree
//     - A
//         - alpha
//             * small_file_0 (< 5 kB)
//         - beta
//             * small_file_1 (< 5 kB)
//         - delta
//             - empty
//         * small_file_2 (< 5 kB)
//         - gamma
//             * small_file_3 (< 5 kB)
//     - B
//         * medium_file (~.3 MB)
//         * big_file (3 MB)
//
// Take special care when modifying this function.  File data depends on order
// and each file size.  If this changes then hashes above
// recording the ipfs import hash tree must be updated manually.
func (sth *ShardingTestHelper) makeTree(t *testing.T) os.FileInfo {
	sth.makeTestFolder(t)
	basepath := sth.path(shardingTestTree)

	// do not re-create
	if folderExists(t, basepath) {
		st, _ := os.Stat(basepath)
		return st
	}

	p0 := shardingTestTree
	paths := [][]string{
		[]string{p0, "A", "alpha"},
		[]string{p0, "A", "beta"},
		[]string{p0, "A", "delta", "empty"},
		[]string{p0, "A", "gamma"},
		[]string{p0, "B"},
	}
	for _, p := range paths {

		makeDir(t, sth.path(p...))
	}

	files := [][]string{
		[]string{p0, "A", "alpha", "small_file_0"},
		[]string{p0, "A", "beta", "small_file_1"},
		[]string{p0, "A", "small_file_2"},
		[]string{p0, "A", "gamma", "small_file_3"},
		[]string{p0, "B", "medium_file"},
		[]string{p0, "B", "big_file"},
	}

	fileSizes := []int{5, 5, 5, 5, 300, 3000}
	for i, fpath := range files {
		path := sth.path(fpath...)
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		sth.randFile(t, f, fileSizes[i])
		f.Close()
	}

	st, err := os.Stat(basepath)
	if err != nil {
		t.Fatal(err)
	}
	return st
}

func (sth *ShardingTestHelper) path(p ...string) string {
	paths := append([]string{shardingTestDir}, p...)
	return filepath.Join(paths...)
}

// Writes randomness to a writer up to the given size (in kBs)
func (sth *ShardingTestHelper) randFile(t *testing.T, w io.Writer, kbs int) {
	buf := make([]byte, 1024)
	for i := 0; i < kbs; i++ {
		sth.randSrc.Read(buf) // read 1 kb
		if _, err := w.Write(buf); err != nil {
			t.Fatal(err)
		}
	}
}

// this creates shardingTestFile in the testFolder. It recreates it every
// time.
func (sth *ShardingTestHelper) makeRandFile(t *testing.T, kbs int) os.FileInfo {
	sth.makeTestFolder(t)
	path := sth.path(shardingTestFile)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sth.randFile(t, f, kbs)
	st, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	return st

}

// GetTreeMultiReader creates and returns a MultiFileReader for a testing
// directory tree. Files are pseudo-randomly generated and are always the same.
// Directory structure:
//   - testingTree
//     - A
//         - alpha
//             * small_file_0 (< 5 kB)
//         - beta
//             * small_file_1 (< 5 kB)
//         - delta
//             - empty
//         * small_file_2 (< 5 kB)
//         - gamma
//             * small_file_3 (< 5 kB)
//     - B
//         * medium_file (~.3 MB)
//         * big_file (3 MB)
//
// The total size in ext4 is ~3420160 Bytes = ~3340 kB = ~3.4MB
func (sth *ShardingTestHelper) GetTreeMultiReader(t *testing.T) *files.MultiFileReader {
	sf := sth.GetTreeSerialFile(t)
	slf := files.NewSliceFile("", "", []files.File{sf})
	return files.NewMultiFileReader(slf, true)
}

func (sth *ShardingTestHelper) GetTreeSerialFile(t *testing.T) files.File {
	st := sth.makeTree(t)
	sf, err := files.NewSerialFile(shardingTestTree, sth.path(shardingTestTree), false, st)
	if err != nil {
		t.Fatal(err)
	}
	return sf
}

// GetRandFileMultiReader creates and returns a MultiFileReader for
// a testing random file of the given size (in kbs). The random
// file is different every time.
func (sth *ShardingTestHelper) GetRandFileMultiReader(t *testing.T, kbs int) *files.MultiFileReader {
	st := sth.makeRandFile(t, kbs)
	sf, err := files.NewSerialFile("randomfile", sth.path(shardingTestFile), false, st)
	if err != nil {
		t.Fatal(err)
	}
	slf := files.NewSliceFile("", "", []files.File{sf})
	return files.NewMultiFileReader(slf, true)
}

// Clean deletes any folder and file generated by this helper.
func (sth *ShardingTestHelper) Clean() {
	os.RemoveAll(shardingTestDir)
}
