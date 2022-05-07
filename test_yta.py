import ctypes

dynload = ctypes.CDLL("./ytarchive.so")
dynload.initialize.argtypes = [ctypes.c_char_p]
dynload.initialize.restype = ctypes.c_long
print(dynload.initialize(b"11111111111"))
