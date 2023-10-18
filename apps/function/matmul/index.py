import numpy as np


def matmul(n):
    A = np.random.rand(n, n)
    B = np.random.rand(n, n)
    C = np.matmul(A, B)


def handler(event, context={}):
    n = len(event.get('n'))
    matmul(n)
    return {'msg': "Finish computing."}
