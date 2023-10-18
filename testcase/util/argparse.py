import argparse
import ast


def init_args(params):
    arg_parser = argparse.ArgumentParser()
    arg_parser = add_argument_base(arg_parser)
    args = arg_parser.parse_args(params)
    return args

def add_argument_base(arg_parser):
    arg_parser.add_argument('--srcPath', default="")
    arg_parser.add_argument('--memory', type=int, default=128)
    arg_parser.add_argument('--param', default={}, type=ast.literal_eval, help='')
    arg_parser.add_argument('--appName',  default="", help='')
    arg_parser.add_argument('--cliBase',  default="", help='')
    arg_parser.add_argument('--testCaseDir',  default="",  help='')
    arg_parser.add_argument('--provider',  default="", help='')
    arg_parser.add_argument('--resultDir',  default="", help='')
    arg_parser.add_argument('--srcPathList',  default="", help='')
    arg_parser.add_argument('--reqPathList',  default="", help='')
    arg_parser.add_argument('--memSizeList',  default="", help='')
    arg_parser.add_argument('--stageNameList',  default="", help='')
    return arg_parser