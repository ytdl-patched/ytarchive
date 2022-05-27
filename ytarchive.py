import ctypes

dynload = ctypes.CDLL("./_ytarchive.so")
dynload.goInfo.argtypes = []
dynload.goInfo.restype = ctypes.c_char_p
up_siz_str, = dynload.goInfo().decode().split(" ")
up_size = int(up_siz_str)

# find a corresponding type for uintptr in ctypes
for tp in (ctypes.c_uint, ctypes.c_ulong, ctypes.c_ulonglong):
    if ctypes.sizeof(tp) == up_size:
        go_uintptr = tp
        break
else:
    import warnings
    warnings.warn(ResourceWarning("Failed to figure out corresponding type for uintptr in ctypes. Assuming unsigned long long."))
    go_uintptr = ctypes.c_ulonglong

dynload.initialize.argtypes = [ctypes.c_char_p]
dynload.initialize.restype = go_uintptr

dynload.registerFormat.argtypes = [go_uintptr, ctypes.c_char_p, ctypes.c_char_p, ctypes.c_char_p, ctypes.c_char_p]
dynload.registerFormat.restype = None

dynload.loadCookies.argtypes = [ctypes.c_char_p]
# golang has no corresponding type to _Bool. the returned value can be used in if statements without conversion
dynload.loadCookies.restype = ctypes.c_int

dynload.runDownloader.argtypes = [go_uintptr]
dynload.runDownloader.restype = None

dynload.interrupt.argtypes = [go_uintptr]
dynload.interrupt.restype = None

# timeout argument (which is the 2nd argument) is in milliseconds, unlike time.sleep
dynload.poll.argtypes = [go_uintptr, ctypes.c_int]
dynload.poll.restype = ctypes.c_char_p
