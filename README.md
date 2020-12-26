# check-md5-go
Multithreaded and algoritmically faster Golang script to compare md5 checksum of two lists of files

If you're here maybe you come from https://github.com/DarkFighterLuke/check-md5-ps (or you're simply lucky).

# Prerequisites
You need md5deep to create dump files compatible with this script. You can find it here for Windows, Linux and Mac OS: http://md5deep.sourceforge.net/ .

# Scope
You will need to dump checksums of the two files (or lists of files) into two text files. You can now pass them to the script and it will find all the files which have a md5 checksum mismatch and it will write it down to the specified results file.

At the end of the execution you will know all the corrupted files and the copy them again.

# Usage
The script requires 3 parameters:
  ```
  1. First checksum file
  2. Second checksum file
  3. File where to write negative results of the comparison
 ```
 
The script is equiped with a resume mechanism, so it will start from the last occurence when you interrupted the execution.

# Implementation
This script uses the mergesort algorithm to reorder lines in the second file passed. This allows to use a binary search algorithm to quickly find matches.
Further, the mergesort algorithm implementation used is multithread, allowing better performances on multi core processors.
