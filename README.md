# nyan
Yet another netcat for fast file transfer

When I need to transfer a file in safe environment (e.g. LAN / VMs),
I just want to use a simple command without cumbersome authentication, GUI, server etc.
`ncat` usually works very well in this case.
However, lastest ncat is super slow on Windows (~32KB/s), and that's why `nyan` was born.
To my surprise, such a naive implementation works very well, even on linux.

## Features
* Plain TCP stream
    * you don't need to use `nyan` at both end
* Progress indicator: Percentage, Size, ETA, Speed

## Speed
Testing commands:
```sh
Linux ncat:
$ ncat -lvp 1234 --send-only < /dev/zero
$ ncat 127.0.0.1 1234 --recv-only | pv -perb > /dev/null

Linux nyan:
$ nyan send /dev/zero 1234
$ nyan recv /dev/null 1234 127.0.0.1

Windows:
$ zeros | ncat -lvp 1234 --send-only
$ ncat 127.0.0.1 1234 --recv-only > NUL

Windows nyan:
$ zeros | nyan send - 1234
$ nyan recv NUL 1234 127.0.0.1
```

Testing environment:
* Arch Linux host
* Windows 10 20H2 KVM guest with virtio netif connected to local bridge
* Arch Linux KVM guest with virtio netif connected to local bridge
* [Ncat 5.59BETA1](https://nmap.org/ncat/) on Windows performs better than latest version.

| A          | B           | nyan (A->B) | ncat (A->B) | nyan (B->A) | ncat (B->A) |
|:----------:|:-----------:|:-----------:|:-----------:|:-----------:|:-----------:|
| Linux Host | localhost   |    6.2 GB/s |    2.6 GB/s |    -        |    -        |
| Linux VM   | localhost   |    6.2 GB/s |    1.9 GB/s |    -        |    -        |
| Windows VM | localhost   |    2.9 GB/s |    0.6 GB/s |    -        |    -        |
| Linux Host | Linux VM    |    0.8 GB/s |    0.7 GB/s |    5.0 GB/s |    2.2 GB/s |
| Linux Host | Windows VM  |    3.2 GB/s |    1.2 GB/s |    2.6 GB/s |    0.3 GB/s |
| Linux VM   | Windows VM  |    2.5 GB/s |    0.8 GB/s |    0.6 GB/s |    0.3 GB/s |
