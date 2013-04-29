package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func findPackagesInDir( dirname string,	package_dirs_by_name map[string] string ) {
	dir, err := os.Open(dirname)
	if err != nil {
		return
	}
	entries, err := dir.Readdir(0)
	dir.Close()
	if err != nil {
		return
	}
	// If there's a manifest.xml in the directory, it is a package
	// and we can stop recursing.
	for _, entry := range entries {
		if entry.Name() == "manifest.xml" {
			package_name := path.Base( dirname )
			_, package_found_already := package_dirs_by_name[ package_name ]
			if !package_found_already {
				package_dirs_by_name[ package_name ] = dirname
			}
			return
		}
	}
	// If no manifest.xml, must not be a package directory, so
	// recurse into each subdir.
	for _, entry := range entries {
		if entry.IsDir() {
			full_entry := path.Join( dirname, entry.Name() )
			findPackagesInDir( full_entry, package_dirs_by_name )
		}
	}
}

func listPackagesInPath( path string ) map[string]string {
	package_dirs_by_name := make( map[string] string )
	directories := filepath.SplitList( path )
	for _, dir := range directories {
		dir, err := filepath.Abs( dir )
		if( err == nil ) {
			findPackagesInDir( dir, package_dirs_by_name )
		}
	}
	return package_dirs_by_name
}

func generatePackageMessages( output_base_dir string, pkg_name string, pkg_dir string ) {
	msg_dir := path.Join( pkg_dir, "msg" )
	dir, err := os.Open( msg_dir )
	if err != nil {
		return
	}
	entries, err := dir.Readdir(0)
	dir.Close()
	if err != nil {
		return
	}

	output_dir := path.Join( output_base_dir, pkg_name )
	err = os.MkdirAll( output_dir, 0755 )
	if err != nil {
		fmt.Printf( "Failed to create output dir %v.  Error: %v\n", output_dir, err )
		return
	}

	for _, entry := range entries {
		if path.Ext( entry.Name() ) == ".msg" {
			msg_file := path.Join( msg_dir, entry.Name() )
			fmt.Printf( "Processing %v.\n", msg_file )
			processMessageFile( output_dir, pkg_name, msg_file, entry )
		}
	}
}

func filenameWithoutExtension( name string ) string {
	lastDot := strings.LastIndex( name, "." )
	if lastDot == -1 {
		return name
	}
	return name[:lastDot]
}

func processMessageFile(
	output_dir string,
	pkg_name string,
	msg_file string,
	msg_file_info os.FileInfo ) {

	in, err := os.Open( msg_file )
	if( err != nil ) {
		fmt.Printf( "Failed to open message file %v.  Error: %v\n", msg_file, err )
		return
	}

	message_name := filenameWithoutExtension( msg_file_info.Name() )

	out_file := path.Join( output_dir, message_name + ".go" )
	out, err := os.Create( out_file )
	if( err != nil ) {
		fmt.Printf( "Failed to open output file %v.  Error: %v\n", out_file, err )
		return
	}

	processFile( pkg_name, message_name, in, out )
}

func main() {
	output_base_dir := path.Join( os.Getenv( "HOME" ), ".ros", "msg_gen", "go", "src" )
	err := os.MkdirAll( output_base_dir, 0755 )
	if( err != nil ) {
		fmt.Printf( "Failed to create output directory %v.  Error: %v\n",
			output_base_dir, err )
		return
	}
	package_dirs_by_name := listPackagesInPath( os.Getenv( "ROS_PACKAGE_PATH" ))
	for name, dir := range package_dirs_by_name {
		fmt.Printf( "Processing package %v in dir %v\n", name, dir )
		generatePackageMessages( output_base_dir, name, dir )
	}
}

// Generated message code will go in:
// ~/.ros/msg_gen/go/src/package_name/*.go

