EXEQ (WIP)
===========
> Execute shell commands in queues via cli or http interface using redis as a broker.

Why
===
> I'm utilizing background jobs heavily in my projects especially crawling projects, let's say we have a scrapy project and we want to make its spiders run on a distributed queue workers without any complex setup, sod all what I need is to call `exeq` and pass scrapy commands to it either from a `cli` or its `http` api.