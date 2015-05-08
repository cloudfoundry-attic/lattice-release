#!/usr/bin/ruby
require "json"

## TODO: Generate JSON from hidden metadata in markdown docs.

lattice_path = ARGV.shift
website_path = ARGV.shift

raise "missing lattice_path argument" unless lattice_path 
raise "missing website_path argument" unless website_path 

# Read JSON from a file, iterate over objects
file = open("#{lattice_path}/docs/docs-metadata.json")
json = file.read

jsonMetadata = JSON.parse(json)

jsonMetadata.each do |doc|
  docFile = open("#{website_path}/middleman/source/docs/#{doc[0]}", 'w')

  docFile.write("---\n")
  doc[1].each do |key, value|
  	docFile.write("#{key}: #{value}\n")
  end
  docFile.write("---\n\n")

  docFile.write(File.read("#{lattice_path}/docs/#{doc[0]}"))

  docFile.close
end


