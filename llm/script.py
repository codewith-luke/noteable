#!/usr/bin/env python3
import sys
import getopt
import main


def myfunc(argv):
    arg_input = ""
    arg_help = "{0} -i <input>".format(argv[0])

    try:
        opts, args = getopt.getopt(argv[1:], "hi:u:o:", ["help", "input="])

    except:
        print(arg_help)
        sys.exit(2)

    for opt, arg in opts:
        if opt in ("-h", "--help"):
            print(arg_help)  # print the help message
            sys.exit(2)
        elif opt in ("-i", "--input"):
            arg_input = arg

    print("debug: handling message")
    main.call(arg_input)


if __name__ == "__main__":
    myfunc(sys.argv)
