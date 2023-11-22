package protoc

import (
	"fmt"
	"strings"

	"github.com/aserto-dev/clui"
	"github.com/aserto-dev/mage-loot/deps"
	"github.com/magefile/mage/sh"
)

type protocArgs struct {
	args       []string
	protoFiles []string
}

// Arg represents a protoc CLI argument.
type Arg func(*protocArgs)

var (
	ui = clui.NewUI()
)

// Run runs the protoc CLI.
func Run(args ...Arg) error {
	protocArgs := &protocArgs{}

	for _, arg := range args {
		arg(protocArgs)
	}

	finalArgs := []string{}

	finalArgs = append(finalArgs, protocArgs.args...)
	finalArgs = append(finalArgs, protocArgs.protoFiles...)

	ui.Normal().
		WithStringValue("command", "protoc "+strings.Join(protocArgs.args, "\n")).
		Msg(">>> executing protoc")

	return sh.RunV(deps.BinPath("protoc"), finalArgs...)
}

// Add adds a new "name=value" style argument (e.g. --proto_path=./foo).
func Add(name, value string) func(*protocArgs) {
	return AddArg(fmt.Sprintf("%s=%s", name, value))
}

// AddOpt adds a plugin option argument (e.g. --go_opt=paths=source_relative).
func AddOpt(name, opt, value string) func(*protocArgs) {
	return AddArg(fmt.Sprintf("%s=%s=%s", name, opt, value))
}

// AddArg adds a simple argument (e.g. --deterministic_output).
func AddArg(arg string) func(*protocArgs) {
	return func(o *protocArgs) {
		o.args = append(o.args, arg)
	}
}

// ProtoFile adds a new input file.
func ProtoFile(file string) func(*protocArgs) {
	return func(o *protocArgs) {
		o.protoFiles = append(o.protoFiles, file)
	}
}

func ProtoFiles(files []string) func(*protocArgs) {
	return func(o *protocArgs) {
		o.protoFiles = append(o.protoFiles, files...)
	}
}

// Version - Show version info and exit.
func Version() Arg { return AddArg("--version") }

// Help - Show this text and exit.
func Help() Arg { return AddArg("--help") }

// DeterministicOutput - When using --encode, ensure map fields are
// deterministically ordered. Note thatthis order is not
// canonical, and changes across builds or releases of protoc.
func DeterministicOutput() Arg { return AddArg("--deterministic_output") }

// DecodeRaw - Read an arbitrary protocol message from
// standard input and write the raw tag/value
// pairs in text format to standard output.  No
// PROTO_FILES should be given when using this
// flag.
func DecodeRaw() Arg { return AddArg("--decode_raw") }

// Out - output file.
func Out(file string) Arg { return AddArg("-o" + file) }

// IncludeImports - When using --descriptor_set_out, also include
// all dependencies of the input files in the
// set, so that the set is self-contained.
func IncludeImports() Arg { return AddArg("--include_imports") }

// IncludeSourceInfo - When using --descriptor_set_out, do not strip
// SourceCodeInfo from the FileDescriptorProto.
// This results in vastly larger descriptors that
// include information about the original
// location of each decl in the source file as
// well as surrounding comments.
func IncludeSourceInfo() Arg { return AddArg("--include_source_info") }

// PrintFreeFieldNumbers - Print the free field numbers of the messages
// defined in the given proto files. Groups share
// the same field number space with the parent
// message. Extension ranges are counted as
// occupied fields numbers.
func PrintFreeFieldNumbers() Arg { return AddArg("--print_free_field_numbers") }

// ProtoPath - Specify the directory in which to search for
// imports.  May be specified multiple times;
// directories will be searched in order.  If not
// given, the current working directory is used.
// If not found in any of the these directories,
// the --descriptor_set_in descriptors will be
// checked for required proto file.
func ProtoPath(path string) Arg { return Add("--proto_path", path) }

// Encode - Read a text-format message of the given type
// from standard input and write it in binary
// to standard output.  The message type must
// be defined in PROTO_FILES or their imports.
func Encode(messageType string) Arg { return Add("--encode", messageType) }

// Decode - Read a binary message of the given type from
// standard input and write it in text format
// to standard output.  The message type must
// be defined in PROTO_FILES or their imports.
func Decode(messageType string) Arg { return Add("--decode", messageType) }

// DescriptorSetIn - Specifies a delimited list of FILES
// each containing a FileDescriptorSet (a
// protocol buffer defined in descriptor.proto).
// The FileDescriptor for each of the PROTO_FILES
// provided will be loaded from these
// FileDescriptorSets. If a FileDescriptor
// appears multiple times, the first occurrence
// will be used.
func DescriptorSetIn(files string) Arg { return Add("--descriptor_set_in", files) }

// DescriptorSetOut - Writes a FileDescriptorSet (a protocol buffer,
// defined in descriptor.proto) containing all of
// the input files to FILE.
func DescriptorSetOut(file string) Arg { return Add("--descriptor_set_out=%s", file) }

// DependencyOut - Write a dependency output file in the format
// expected by make. This writes the transitive
// set of input file paths to FILE.
func DependencyOut(file string) Arg { return Add("--dependency_out", file) }

// ErrorFormat - Set the format in which to print errors.
// FORMAT may be 'gcc' (the default) or 'msvs'
// (Microsoft Visual Studio format).
func ErrorFormat(format string) Arg { return Add("--error_format", format) }

// Plugin - Specifies a plugin executable to use.
// Normally, protoc searches the PATH for
// plugins, but you may specify additional
// executables not in the path using this flag.
// Additionally, EXECUTABLE may be of the form
// NAME=PATH, in which case the given plugin name
// is mapped to the given executable even if
// the executable's own name differs.
func Plugin(executable string) Arg { return Add("--plugin", executable) }

// CPPOut - Generate C++ header and source.
func CPPOut(outDir string) Arg { return Add("--cpp_out", outDir) }

// CSharpOut - Generate C# source file.
func CSharpOut(outDir string) Arg { return Add("--csharp_out", outDir) }

// JavaOut - Generate Java source file.
func JavaOut(outDir string) Arg { return Add("--java_out", outDir) }

// JSOut - Generate JavaScript source.
func JSOut(outDir string) Arg { return Add("--js_out", outDir) }

// ObjCOut - Generate Objective-C header and source.
func ObjCOut(outDir string) Arg { return Add("--objc_out", outDir) }

// PHPOut - Generate PHP source file.
func PHPOut(outDir string) Arg { return Add("--php_out", outDir) }

// PythonOut - Generate Python source file.
func PythonOut(outDir string) Arg { return Add("--python_out", outDir) }

// RubyOut - Generate Ruby source file.
func RubyOut(outDir string) Arg { return Add("--ruby_out", outDir) }

// GoOut - Generate go source files.
func GoOut(outDir string) Arg { return Add("--go_out", outDir) }

// GoGrpcOut - Generate go grpc files.
func GoGrpcOut(outDir string) Arg { return Add("--go-grpc_out", outDir) }

// GrpcGatewayOut - Generate a go grpc open api gateway.
func GrpcGatewayOut(outDir string) Arg { return Add("--grpc-gateway_out", outDir) }

// OpenAPIV2Out - Generate an Open API v2 definition.
func OpenAPIV2Out(outDir string) Arg { return Add("--openapiv2_out", outDir) }

// DocOut - Generate docs.
func DocOut(outDir string) Arg { return Add("--doc_out", outDir) }

// DocOpt - Doc plugin option.
func DocOpt(value string) Arg { return Add("--doc_opt", value) }

// GoOpt - Go plugin option.
func GoOpt(opt, value string) Arg { return AddOpt("--go_opt", opt, value) }

// GoGrpcOpt - Go grpc plugin option.
func GoGrpcOpt(opt, value string) Arg { return AddOpt("--go-grpc_opt", opt, value) }

// GrpcGatewayOpt - Grpc gateway plugin option.
func GrpcGatewayOpt(opt, value string) Arg { return AddOpt("--grpc-gateway_opt", opt, value) }

// OpenAPIV2Opt - Open Api plugin option.
func OpenAPIV2Opt(opt, value string) Arg { return AddOpt("--openapiv2_opt", opt, value) }

// CSharpOpt - C# Options.
func CSharpOpt(opt, value string) Arg { return AddOpt("--csharp_opt", opt, value) }
