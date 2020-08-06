package http

import (
	"github.com/stretchr/testify/assert"
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
	got, _ := trie.retrieve("a/b/c", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "walla", got)
	got, _ = trie.retrieve("a/d/g", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "kuku", got)
	got, _ = trie.retrieve("x/b/c", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "lala", got)
	got, _ = trie.retrieve("a/x/b", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "one", got)
	got, _ = trie.retrieve("a/b/d", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "two", got)

	got, _ = trie.retrieve("a/b", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, nil, got)
	got, _ = trie.retrieve("a/b/c/d", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, nil, got)
	got, _ = trie.retrieve("g/t/x", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "three", got)

	got, params := trie.retrieve("index1/insert/12", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "bingo", got)
	assert.Equal(t, 2, len(params))
	assert.Equal(t, "index1", params["index"])
	assert.Equal(t, "12", params["docId"])
}

func TestEmptyPath(t *testing.T) {
	// Arrange
	trie := newPathTrie()
	trie.insert("/", "walla")

	// Action, Assert
	got, _ := trie.retrieve("/", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "walla", got)
	got, _ = trie.retrieve("", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "walla", got)
}

func TestDifferentNamesOnDifferentPath(t *testing.T) {
	// Arrange
	trie := newPathTrie()
	trie.insert("/a/{type}", "test1")
	trie.insert("/b/{name}", "test2")

	// Action, Assert
	got, params := trie.retrieve("/a/test", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "test1", got)
	assert.Equal(t, "test", params["type"])

	got, params = trie.retrieve("/b/testX", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "test2", got)
	assert.Equal(t, "testX", params["name"])
}

func TestSameNameOnDifferentPath(t *testing.T) {
	// Arrange
	trie := newPathTrie()
	trie.insert("/a/c/{name}", "test1")
	trie.insert("/b/{name}", "test2")

	// Action, Assert
	got, params := trie.retrieve("/a/c/test", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "test1", got)
	assert.Equal(t, "test", params["name"])

	got, params = trie.retrieve("/b/test", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "test2", got)
	assert.Equal(t, "testX", params["name"])
}

func TestPreferNonWildcardExecution(t *testing.T) {
	// Arrange
	trie := newPathTrie()
	trie.insert("{test}", "test1")
	trie.insert("b", "test2")
	trie.insert("{test}/a", "test3")
	trie.insert("b/a", "test4")
	trie.insert("{test}/{testB}", "test5")
	trie.insert("{test}/x/{testC}", "test6")

	// Action, ASsert
	got, _ := trie.retrieve("/b", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "test2", got)
	got, _ = trie.retrieve("/b/a", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "test4", got)
	got, _ = trie.retrieve("/v/x", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "test5", got)
	got, _ = trie.retrieve("/v/x/c", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "test6", got)
}

func TestWildcardMatchingModes(t *testing.T) {
	// Arrange
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

	// Action, Assert
	got, _ := trie.retrieve("/a", EXPLICIT_NODES_ONLY)
	assert.Equal(t, nil, got)
	got, _ = trie.retrieve("/a", WILDCARD_ROOT_NODES_ALLOWED)
	assert.Equal(t, "test1", got)
	got, _ = trie.retrieve("/a", WILDCARD_LEAF_NODES_ALLOWED)
	assert.Equal(t, "test1", got)
	got, _ = trie.retrieve("/a", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "test1", got)

	allPaths := trie.retrieveAll("/a")
	got, _, _ = allPaths()
	assert.Equal(t, nil, got)
	got, _, _ = allPaths()
	assert.Equal(t, "test1", got)
	got, _, _ = allPaths()
	assert.Equal(t, "test1", got)
	got, _, _ = allPaths()
	assert.Equal(t, "test1", got)
	_, _, err := allPaths()
	assert.NotNil(t, err)

	got, _ = trie.retrieve("/a/b/c", EXPLICIT_NODES_ONLY)
	assert.Equal(t, nil, got)
	got, _ = trie.retrieve("/a/b/c", WILDCARD_ROOT_NODES_ALLOWED)
	assert.Equal(t, "test5", got)
	got, _ = trie.retrieve("/a/b/c", WILDCARD_LEAF_NODES_ALLOWED)
	assert.Equal(t, "test7", got)
	got, _ = trie.retrieve("/a/b/c", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "test7", got)

	allPaths = trie.retrieveAll("/a/b/c")
	got, _, _ = allPaths()
	assert.Equal(t, nil, got)
	got, _, _ = allPaths()
	assert.Equal(t, "test5", got)
	got, _, _ = allPaths()
	assert.Equal(t, "test7", got)
	got, _, _ = allPaths()
	assert.Equal(t, "test7", got)
	_, _, err = allPaths()
	assert.NotNil(t, err)

	got, _ = trie.retrieve("/x/y/z", EXPLICIT_NODES_ONLY)
	assert.Equal(t, nil, got)
	got, _ = trie.retrieve("/x/y/z", WILDCARD_ROOT_NODES_ALLOWED)
	assert.Equal(t, nil, got)
	got, _ = trie.retrieve("/x/y/z", WILDCARD_LEAF_NODES_ALLOWED)
	assert.Equal(t, nil, got)
	got, _ = trie.retrieve("/x/y/z", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "test9", got)

	allPaths = trie.retrieveAll("/x/y/z")
	got, _, _ = allPaths()
	assert.Equal(t, nil, got)
	got, _, _ = allPaths()
	assert.Equal(t, nil, got)
	got, _, _ = allPaths()
	assert.Equal(t, nil, got)
	got, _, _ = allPaths()
	assert.Equal(t, "test9", got)
	_, _, err = allPaths()
	assert.NotNil(t, err)

	got, _ = trie.retrieve("/d/e/f", EXPLICIT_NODES_ONLY)
	assert.Equal(t, nil, got)
	got, _ = trie.retrieve("/d/e/f", WILDCARD_ROOT_NODES_ALLOWED)
	assert.Equal(t, nil, got)
	got, _ = trie.retrieve("/d/e/f", WILDCARD_LEAF_NODES_ALLOWED)
	assert.Equal(t, nil, got)
	got, _ = trie.retrieve("/d/e/f", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "test10", got)

	allPaths = trie.retrieveAll("/d/e/f")
	got, _, _ = allPaths()
	assert.Equal(t, nil, got)
	got, _, _ = allPaths()
	assert.Equal(t, nil, got)
	got, _, _ = allPaths()
	assert.Equal(t, nil, got)
	got, _, _ = allPaths()
	assert.Equal(t, "test10", got)
	_, _, err = allPaths()
	assert.NotNil(t, err)

}

func TestExplicitMatchingMode(t *testing.T) {
	// Arrange
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

	// Action, Assert
	got, _ := trie.retrieve("/a", EXPLICIT_NODES_ONLY)
	assert.Equal(t, "test2", got)
	got, _ = trie.retrieve("/x", EXPLICIT_NODES_ONLY)
	assert.Equal(t, nil, got)
	got, _ = trie.retrieve("/a/b", EXPLICIT_NODES_ONLY)
	assert.Equal(t, "test6", got)
	got, _ = trie.retrieve("/a/x", EXPLICIT_NODES_ONLY)
	assert.Equal(t, nil, got)
	got, _ = trie.retrieve("/a/b/c", EXPLICIT_NODES_ONLY)
	assert.Equal(t, "test10", got)
	got, _ = trie.retrieve("/x/y/z", EXPLICIT_NODES_ONLY)
	assert.Equal(t, nil, got)
}

func TestSamePathConcreteResolution(t *testing.T) {
	// Arrange
	trie := newPathTrie()
	trie.insert("{x}/{y}/{z}", "test1")
	trie.insert("{x}/_y/{k}", "test2")

	// Action, Assert
	got, params := trie.retrieve("/a/b/c", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "test1", got)
	assert.Equal(t, "a", params["x"])
	assert.Equal(t, "b", params["y"])
	assert.Equal(t, "c", params["z"])

	params = map[string]string{}
	got, params = trie.retrieve("/a/_y/c", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "test2", got)
	assert.Equal(t, "a", params["x"])
	assert.Equal(t, "c", params["k"])
}

func TestNamedWildcardAndLookupWithWildcard(t *testing.T) {
	// Arrange
	trie := newPathTrie()
	trie.insert("x/{test}", "test1")
	trie.insert("{test}/a", "test2")
	trie.insert("/{test}", "test3")
	trie.insert("/{test}/_endpoint", "test4")
	trie.insert("/*/{test}/_endpoint", "test5")

	// Action, Assert
	got, params := trie.retrieve("/x/*", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "test1", got)
	assert.Equal(t, "*", params["test"])

	got, params = trie.retrieve("/b/a", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "test2", got)
	assert.Equal(t, "b", params["test"])

	got, params = trie.retrieve("/8", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "test3", got)
	assert.Equal(t, "8", params["test"])

	got, params = trie.retrieve("/*/_endpoint", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "test4", got)
	assert.Equal(t, "*", params["test"])

	got, params = trie.retrieve("a/*/_endpoint", WILDCARD_NODES_ALLOWED)
	assert.Equal(t, "test5", got)
	assert.Equal(t, "*", params["test"])
}

//func TestEscapedSlashWithinUrl(t *testing.T) {
//	// Arrange
//	trie := newPathTrie()
//	trie.insert("/{index}/{type}/{id}", "test")
//
//	// Action, Assert
//	params := map[string]string{}
//	assert.Equal(t, trie.retrieve("/index/type/a%2Fe", params, WILDCARD_NODES_ALLOWED), "test")
//	assert.Equal(t, params["index"], "index")
//	assert.Equal(t, params["type"], "type")
//	assert.Equal(t, params["id"], "a/e")
//
//	params = map[string]string{}
//	assert.Equal(t, trie.retrieve("/<logstash-{now%2Fd}>/type/id", params, WILDCARD_NODES_ALLOWED), "test")
//	assert.Equal(t, params["index"], "<logstash-{now/d}>")
//	assert.Equal(t, params["type"], "type")
//	assert.Equal(t, params["id"], "id")
//}
