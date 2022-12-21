# Overview

[paepche.de/codereview](https://paepcke.de/codereview)

## Code Review: Save time when track source code changes.

Automatically filter all the noise from commits:
- reformat (whitespace, tabs, blank lines, ...)
- comments (spellfixes, re-wording, copyright, year updates ...)
- reordering (data types, interface arguments, grouping, style fixes ...)

## Compact & Compile: Optimize sourcecode for the compiler.

Shrink size of sourcecode you have to:
- store, transmit, compress, fail to deduplicate
- no more compliler cache trashing caused by noise (ccache, go-build, ...)
- parses full AST tree and prints a very compact, optimized version (golang)

# Showtime Code Compact

Compact source code before (compile|store|review)

- example : freebsd/freebsd-src mixed sourcecode
- process : 1.4GB, 99.818 files, 21.364.953 LOC
- save    : 6.293.356 LOC removed (~171 MB) 
- time    : > less than a second (x86/4core/2.2GHz)

```Shell
git.checkout freebsd && codereview --inplace .
CODEREVIEW [start] [/tmp/freebsd] 
CODEREVIEW [_done] [902.557426ms]
CODEREVIEW [stats] [total files: 99818] [processed: 50343]
CODEREVIEW [stats] [total loc: 21363953] [removeable: 6293365] [savings: 171.1 Mbytes]
[...]
```
# Showtime Code Review 

## Full Repo View: 

```Shell
codereview .
CODEREVIEW [start] [/codereview] 
[d3d35f5cca65b03f8bd09e0bc86f2ee700b939583e8018fbb6cc9728ccae4f02] [/codereview/.bootstrap.sh]
[0da6539826e85a2766c38807663ad1f102712684a02de57d3fc3064e12014cc5] [/codereview/generic.go]
[7ab41391aa3320d84fa99f55ebacaf111dc695ae0bb37cfbe04a27cd93085185] [/codereview/io.go]
[4139fde043f804d47c7bca1a791748232100bf69176842f1a32216a9a75e3c80] [/codereview/core.go]
[8ea982d9269ec348f593f73389233b49743fc0f3548e683ef1c4651eac4422d9] [/codereview/.archive/worker_vim.go]
[c823fd27dd3cc351b9ae320dca271af9bdb1024e2906828ee33baa1f6684b195] [/codereview/api.go]
[6220b152fcee96dad3084d84f1bd907150d2fabb19c23d63f0d7218ece608d3c] [/codereview/worker_asm.go]
[6c28ea03823b5148a9bb6c028583d16903e133602bb737978b2de35a35d0448b] [/codereview/walker.go]
[5c9dfcea7547c4492b826947961c006c38872894c0a8dbb229a2a593a038fb25] [/codereview/worker_sh.go]
[554044178a68de08cd16125ab57565230462b8225434046c50f20156d29b4577] [/codereview/worker_go.go]
[15ced74c8fa0dfde9c4f609e7b121ea3bbb806c11cd6885ed65f50ab6df84921] [/codereview/worker_c.go]
[39b58a23f0ae0a4497b33ede4dc67943f4c1595b468243a18b0a8a09428ae471] [/codereview/worker_make.go]
[bb412f445706395851c594cafcc1e5f20376fd4743410fc7aafa08f74566b8b7] [/codereview/cmd/codereview/main.go]
CODEREVIEW [_done] [168.329114ms]
CODEREVIEW [stats] [total files: 49] [processed: 19]
CODEREVIEW [stats] [total loc: 2289] [removeable: 191] [savings: 3.2 kbytes]
```

## Code Review Example 

Produce and use Codereview Hashes of individual files, that does not change via 'noise'.

```Shell
# example: review code changes of api.go
sha256sum api.go                 	# get classic hash
	result: 7064f2b238668b830877f2ee3db582d291017e11bd31479662758ed46072f50f 
codereview --hash api.go                # get codereview hash
	result: c823fd27dd3cc351b9ae320dca271af9bdb1024e2906828ee33baa1f6684b195


# make a style change (replace all tabs with whitespaces)
sed -i '' -e 's/\t/   /g' api.go 	# make an noisy change, that does not impact code generation
sha256sum api.go		 	# see changed classic hash changed 
	result: 67c5e0c746150a022c9a48a4c759ac97221e9710f589379fe861b1418b8a4351  api.go
codereview --hash api.go                # see codereview hash, not impressed by style change, skip manual review
	result: c823fd27dd3cc351b9ae320dca271af9bdb1024e2906828ee33baa1f6684b195


# make a change that changes resulting binary executeable 
sed -i '' -e 's/true/false/g' api.go	# make an change with impact to code generation 
sha256sum api.go		 	# see changed classic hash changed 
	result: d3ad85e4831c096f825883c6fd9525b8696ba24ef5ff39c2aec0e91199bc1367  api.go
codereview --hash api.go                # see coderevieview hash changed as well, so change might needs manual review
	result: 877e7b5fedc48d8547bb8c34bbbe4e3e8efc44e7c9bc30b2a95ffd18f796f8a9
```


## More?

```Shell
codereview --help
syntax: codereview [options] <file|directory>

--inplace [-w]
		write changes direct back to source files
		without this option changes outputs goes to stdout

--hash [-H]
		show hashsum only

--exclude [-e]
		exclude all directories matching any of the keywords
		this option can be specified several times

--disable-c
--disable-go
--disable-sh
--disable-asm
--disable-make
--disable-hidden-files
		disable [files|language support] in recursive directory walk

--verbose [-v]
--silent [-q]
--debug [-d]
--help [-h]

```
