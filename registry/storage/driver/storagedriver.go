package driver

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/docker/distribution/context"
)

// Version is a string representing the storage driver version, of the form
// Major.Minor.
// The registry must accept storage drivers with equal major version and greater
// minor version, but may not be compatible with older storage driver versions.
type Version string

// Major returns the major (primary) component of a version.
func (version Version) Major() uint {
	majorPart := strings.Split(string(version), ".")[0]
	major, _ := strconv.ParseUint(majorPart, 10, 0)
	return uint(major)
}

// Minor returns the minor (secondary) component of a version.
func (version Version) Minor() uint {
	minorPart := strings.Split(string(version), ".")[1]
	minor, _ := strconv.ParseUint(minorPart, 10, 0)
	return uint(minor)
}

// CurrentVersion is the current storage driver Version.
const CurrentVersion Version = "0.1"

// StorageDriver defines methods that a Storage Driver must implement for a
// filesystem-like key/value object storage.
type StorageDriver interface {
	// Name returns the human-readable "name" of the driver, useful in error
	// messages and logging. By convention, this will just be the registration
	// name, but drivers may provide other information here.
	Name() string

	// GetContent retrieves the content stored at "path" as a []byte.
	// This should primarily be used for small objects.
	GetContent(ctx context.Context, path string) ([]byte, error)

	// PutContent stores the []byte content at a location designated by "path".
	// This should primarily be used for small objects.
	PutContent(ctx context.Context, path string, content []byte) error

	// ReadStream retrieves an io.ReadCloser for the content stored at "path"
	// with a given byte offset.
	// May be used to resume reading a stream by providing a nonzero offset.
	ReadStream(ctx context.Context, path string, offset int64) (io.ReadCloser, error)

	// WriteStream stores the contents of the provided io.ReadCloser at a
	// location designated by the given path.
	// May be used to resume writing a stream by providing a nonzero offset.
	// The offset must be no larger than the CurrentSize for this path.
	WriteStream(ctx context.Context, path string, offset int64, reader io.Reader) (nn int64, err error)

	// Stat retrieves the FileInfo for the given path, including the current
	// size in bytes and the creation time.
	Stat(ctx context.Context, path string) (FileInfo, error)

	// List returns a list of the objects that are direct descendants of the
	//given path.
	List(ctx context.Context, path string) ([]string, error)

	// Move moves an object stored at sourcePath to destPath, removing the
	// original object.
	// Note: This may be no more efficient than a copy followed by a delete for
	// many implementations.
	Move(ctx context.Context, sourcePath string, destPath string) error

	// Delete recursively deletes all objects stored at "path" and its subpaths.
	Delete(ctx context.Context, path string) error

	// URLFor returns a URL which may be used to retrieve the content stored at
	// the given path, possibly using the given options.
	// May return an ErrUnsupportedMethod in certain StorageDriver
	// implementations.
	URLFor(ctx context.Context, path string, options map[string]interface{}) (string, error)
}

// PathRegexp is the regular expression which each file path must match. A
// file path is absolute, beginning with a slash and containing a positive
// number of path components separated by slashes, where each component is
// restricted to lowercase alphanumeric characters or a period, underscore, or
// hyphen.
var PathRegexp = regexp.MustCompile(`^(/[A-Za-z0-9._-]+)+$`)

// ErrUnsupportedMethod may be returned in the case where a StorageDriver implementation does not support an optional method.
type ErrUnsupportedMethod struct {
	DriverName string
}

func (err ErrUnsupportedMethod) Error() string {
	return fmt.Sprintf("[%s] unsupported method", err.DriverName)
}

// PathNotFoundError is returned when operating on a nonexistent path.
type PathNotFoundError struct {
	Path       string
	DriverName string
}

func (err PathNotFoundError) Error() string {
	return fmt.Sprintf("[%s] Path not found: %s", err.DriverName, err.Path)
}

// InvalidPathError is returned when the provided path is malformed.
type InvalidPathError struct {
	Path       string
	DriverName string
}

func (err InvalidPathError) Error() string {
	return fmt.Sprintf("[%s] Invalid path: %s", err.DriverName, err.Path)
}

// InvalidOffsetError is returned when attempting to read or write from an
// invalid offset.
type InvalidOffsetError struct {
	Path       string
	Offset     int64
	DriverName string
}

func (err InvalidOffsetError) Error() string {
	return fmt.Sprintf("[%s] Invalid offset: %d for path: %s", err.DriverName, err.Offset, err.Path)
}

// Error is a catch-all error type which captures an error string and
// the driver type on which it occured.
type Error struct {
	DriverName string
	Enclosed   error
}

func (err Error) Error() string {
	return fmt.Sprintf("[%s] %s", err.DriverName, err.Enclosed)
}
