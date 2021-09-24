EXEQ (WIP)
===========
> Execute shell commands in queues via cli or http interface using redis as a broker.

Why
===
> I'm utilizing background jobs heavily in my projects especially crawling projects, let's say we have a scrapy project and we want to make its spiders run on a distributed queue workers without any complex setup, sod all what I need is to call `exeq` and pass scrapy commands to it either from a `cli` or its `http` api.

TODOs
========
- [x] Redis based task queue
- [x] Macros (shortcuts for shell commands)
- [x] Mounting files
- [x] Initutive command line interface 
    - [x] `work`
    - [x] `cmd sh -c "echo hello world"`
    - [x] `macro -m MacroName -a ArgKey1=ArgVal1 ...`
    - [x] `ps`
- [ ] Initutive http interface just like what cli provides.
