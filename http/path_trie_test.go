package http

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestPath(t *testing.T) {
	// Arrange
	trie := newPathTrie()
	trie.insert("/a/b/c", "walla")
	trie.insert("a/d/g", "kuku")
	trie.insert("x/b/c", "lala")
	trie.insert("a/x/*", "one")
	trie.insert("a/b/*", "two")
	trie.insert("*/*/x", "three")
	trie.insert("{index}/insert/{docId}", "bingo")

	// Action, Assert
	assert.Equal(t, trie.retrieve("a/b/c", map[string]string{}, WILDCARD_NODES_ALLOWED), "walla")
	assert.Equal(t, trie.retrieve("a/d/g", map[string]string{}, WILDCARD_NODES_ALLOWED), "kuku")
	assert.Equal(t, trie.retrieve("x/b/c", map[string]string{}, WILDCARD_NODES_ALLOWED), "lala")
	assert.Equal(t, trie.retrieve("a/x/b", map[string]string{}, WILDCARD_NODES_ALLOWED), "one")
	assert.Equal(t, trie.retrieve("a/b/d", map[string]string{}, WILDCARD_NODES_ALLOWED), "two")

	assert.Equal(t, trie.retrieve("a/b", map[string]string{}, WILDCARD_NODES_ALLOWED), nil)
	assert.Equal(t, trie.retrieve("a/b/c/d", map[string]string{}, WILDCARD_NODES_ALLOWED), nil)
	assert.Equal(t, trie.retrieve("g/t/x", map[string]string{}, WILDCARD_NODES_ALLOWED), "three")

	params := map[string]string{}
	assert.Equal(t, trie.retrieve("index1/insert/12", params, WILDCARD_NODES_ALLOWED), "bingo")
	assert.Equal(t, len(params), 2)
	assert.Equal(t, params["index"], "index1")
	assert.Equal(t, params["docId"], "12")
}

func TestEmptyPath(t *testing.T) {
	// Arrange
	trie := newPathTrie()
	trie.insert("/", "walla")

	// Action, Assert
	assert.Equal(t, trie.retrieve("/", map[string]string{}, WILDCARD_NODES_ALLOWED), "walla")
	assert.Equal(t, trie.retrieve("", map[string]string{}, WILDCARD_NODES_ALLOWED), "walla")
}

func TestDifferentNamesOnDifferentPath(t *testing.T) {
	// Arrange
	trie := newPathTrie()
	trie.insert("/a/{type}", "test1")
	trie.insert("/b/{name}", "test2")
}

func TestSameNameOnDifferentPath(t *testing.T) {
	trie := newPathTrie()
	trie.insert("/a/c/{name}", "test1")
	trie.insert("/b/{name}", "test2")
}

func TestPreferNonWildcardExecution(t *testing.T) {
	trie := newPathTrie()
	trie.insert("{test}", "test1")
	trie.insert("b", "test2")
	trie.insert("{test}/a", "test3")
	trie.insert("b/a", "test4")
	trie.insert("{test}/{testB}", "test5")
	trie.insert("{test}/x/{testC}", "test6")
}

func TestWildcardMatchingModes(t *testing.T) {
	trie := newPathTrie()
	trie.insert("{testA}", "test1")
	trie.insert("{testA}/{testB}", "test2")
	trie.insert("a/{testB}", "test3")
	trie.insert("{testA}/b", "test4")
	trie.insert("{testA}/b/c", "test5")
	trie.insert("a/{testB}/c", "test6")
	trie.insert("a/b/{testC}", "test7")
	trie.insert("{testA}/b/{testB}", "test8")
	trie.insert("x/{testB}/z", "test9")
	trie.insert("{testA}/{testB}/{testC}", "test10")
}

func TestExplicitMatchingMode(t *testing.T) {
	trie := newPathTrie()
	trie.insert("{testA}", "test1")
	trie.insert("a", "test2")
	trie.insert("{testA}/{testB}", "test3")
	trie.insert("a/{testB}", "test4")
	trie.insert("{testA}/b", "test5")
	trie.insert("a/b", "test6")
	trie.insert("{testA}/b/{testB}", "test7")
	trie.insert("x/{testA}/z", "test8")
	trie.insert("{testA}/{testB}/{testC}", "test9")
	trie.insert("a/b/c", "test10")
}

func TestSamePathConcreteResolution(t *testing.T) {
	trie := newPathTrie()
	trie.insert("{x}/{y}/{z}", "test1")
	trie.insert("{x}/_y/{k}", "test2")
}

func TestNamedWildcardAndLookupWithWildcard(t *testing.T) {
	trie := newPathTrie()
	trie.insert("x/{test}", "test1")
	trie.insert("{test}/a", "test2")
	trie.insert("/{test}", "test3")
	trie.insert("/{test}/_endpoint", "test4")
	trie.insert("/*/{test}/_endpoint", "test5")
}

func TestEscapedSlashWithinUrl(t *testing.T) {
	trie := newPathTrie()
	trie.insert("/{index}/{type}/{id}", "test")
}
