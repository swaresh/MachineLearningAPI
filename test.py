import random
def getopts(argv):
    opts = {}
    while argv:
        if argv[0][0] == '-':
            opts[argv[0]] = argv[1]
        argv = argv[1:]
    return opts

if __name__ == '__main__':
	from sys import argv
	myargs = getopts(argv)
	from time import sleep
	sleep(0.5)
	print {"i": myargs['--i'], "j": myargs['--j'], "k": myargs['--k'], "image": myargs['--image'], "accuracy": random.random()}