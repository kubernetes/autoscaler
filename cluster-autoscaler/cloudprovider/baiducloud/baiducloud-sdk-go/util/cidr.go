/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"fmt"
	"math/big"
	"net"
)

// Subnet takes a parent CIDR range and creates a subnet within it
// with the given number of additional prefix bits and the given
// network number.
//
// For example, 10.3.0.0/16, extended by 8 bits, with a network number
// of 5, becomes 10.3.5.0/24 .
func Subnet(base *net.IPNet, newBits int, num int) (*net.IPNet, error) {
	ip := base.IP
	mask := base.Mask

	parentLen, addrLen := mask.Size()
	newPrefixLen := parentLen + newBits

	if newPrefixLen > addrLen {
		return nil, fmt.Errorf("insufficient address space to extend prefix of %d by %d", parentLen, newBits)
	}

	maxNetNum := uint64(1<<uint64(newBits)) - 1
	if uint64(num) > maxNetNum {
		return nil, fmt.Errorf("prefix extension of %d does not accommodate a subnet numbered %d", newBits, num)
	}

	return &net.IPNet{
		IP:   insertNumIntoIP(ip, num, newPrefixLen),
		Mask: net.CIDRMask(newPrefixLen, addrLen),
	}, nil
}

// Host takes a parent CIDR range and turns it into a host IP address with
// the given host number.
//
// For example, 10.3.0.0/16 with a host number of 2 gives 10.3.0.2.
func Host(base *net.IPNet, num int) (net.IP, error) {
	ip := base.IP
	mask := base.Mask

	parentLen, addrLen := mask.Size()
	hostLen := addrLen - parentLen

	maxHostNum := uint64(1<<uint64(hostLen)) - 1

	numUint64 := uint64(num)
	if num < 0 {
		numUint64 = uint64(-num) - 1
		num = int(maxHostNum - numUint64)
	}

	if numUint64 > maxHostNum {
		return nil, fmt.Errorf("prefix of %d does not accommodate a host numbered %d", parentLen, num)
	}
	var bitlength int
	if ip.To4() != nil {
		bitlength = 32
	} else {
		bitlength = 128
	}
	return insertNumIntoIP(ip, num, bitlength), nil
}

// AddressRange returns the first and last addresses in the given CIDR range.
func AddressRange(network *net.IPNet) (net.IP, net.IP) {
	// the first IP is easy
	firstIP := network.IP

	// the last IP is the network address OR NOT the mask address
	prefixLen, bits := network.Mask.Size()
	if prefixLen == bits {
		// Easy!
		// But make sure that our two slices are distinct, since they
		// would be in all other cases.
		lastIP := make([]byte, len(firstIP))
		copy(lastIP, firstIP)
		return firstIP, lastIP
	}

	firstIPInt, bits := ipToInt(firstIP)
	hostLen := uint(bits) - uint(prefixLen)
	lastIPInt := big.NewInt(1)
	lastIPInt.Lsh(lastIPInt, hostLen)
	lastIPInt.Sub(lastIPInt, big.NewInt(1))
	lastIPInt.Or(lastIPInt, firstIPInt)

	return firstIP, intToIP(lastIPInt, bits)
}

// AddressCount returns the number of distinct host addresses within the given
// CIDR range.
//
// Since the result is a uint64, this function returns meaningful information
// only for IPv4 ranges and IPv6 ranges with a prefix size of at least 65.
func AddressCount(network *net.IPNet) uint64 {
	prefixLen, bits := network.Mask.Size()
	return 1 << (uint64(bits) - uint64(prefixLen))
}

//VerifyNoOverlap takes a list subnets and supernet (CIDRBlock) and verifies
//none of the subnets overlap and all subnets are in the supernet
//it returns an error if any of those conditions are not satisfied
func VerifyNoOverlap(subnets []*net.IPNet, CIDRBlock *net.IPNet) error {
	firstLastIP := make([][]net.IP, len(subnets))
	for i, s := range subnets {
		first, last := AddressRange(s)
		firstLastIP[i] = []net.IP{first, last}
	}
	for i, s := range subnets {
		if !CIDRBlock.Contains(firstLastIP[i][0]) || !CIDRBlock.Contains(firstLastIP[i][1]) {
			return fmt.Errorf("%s does not fully contain %s", CIDRBlock.String(), s.String())
		}
		for j := 0; j < len(subnets); j++ {
			if i == j {
				continue
			}

			first := firstLastIP[j][0]
			last := firstLastIP[j][1]
			if s.Contains(first) || s.Contains(last) {
				return fmt.Errorf("%s overlaps with %s", subnets[j].String(), s.String())
			}
		}
	}
	return nil
}

// PreviousSubnet returns the subnet of the desired mask in the IP space
// just lower than the start of IPNet provided. If the IP space rolls over
// then the second return value is true
func PreviousSubnet(network *net.IPNet, prefixLen int) (*net.IPNet, bool) {
	startIP := checkIPv4(network.IP)
	previousIP := make(net.IP, len(startIP))
	copy(previousIP, startIP)
	cMask := net.CIDRMask(prefixLen, 8*len(previousIP))
	previousIP = Dec(previousIP)
	previous := &net.IPNet{IP: previousIP.Mask(cMask), Mask: cMask}
	if startIP.Equal(net.IPv4zero) || startIP.Equal(net.IPv6zero) {
		return previous, true
	}
	return previous, false
}

// NextSubnet returns the next available subnet of the desired mask size
// starting for the maximum IP of the offset subnet
// If the IP exceeds the maxium IP then the second return value is true
func NextSubnet(network *net.IPNet, prefixLen int) (*net.IPNet, bool) {
	_, currentLast := AddressRange(network)
	mask := net.CIDRMask(prefixLen, 8*len(currentLast))
	currentSubnet := &net.IPNet{IP: currentLast.Mask(mask), Mask: mask}
	_, last := AddressRange(currentSubnet)
	last = Inc(last)
	next := &net.IPNet{IP: last.Mask(mask), Mask: mask}
	if last.Equal(net.IPv4zero) || last.Equal(net.IPv6zero) {
		return next, true
	}
	return next, false
}

//Inc increases the IP by one this returns a new []byte for the IP
func Inc(IP net.IP) net.IP {
	IP = checkIPv4(IP)
	incIP := make([]byte, len(IP))
	copy(incIP, IP)
	for j := len(incIP) - 1; j >= 0; j-- {
		incIP[j]++
		if incIP[j] > 0 {
			break
		}
	}
	return incIP
}

//Dec decreases the IP by one this returns a new []byte for the IP
func Dec(IP net.IP) net.IP {
	IP = checkIPv4(IP)
	decIP := make([]byte, len(IP))
	copy(decIP, IP)
	decIP = checkIPv4(decIP)
	for j := len(decIP) - 1; j >= 0; j-- {
		decIP[j]--
		if decIP[j] < 255 {
			break
		}
	}
	return decIP
}

func checkIPv4(ip net.IP) net.IP {
	// Go for some reason allocs IPv6len for IPv4 so we have to correct it
	if v4 := ip.To4(); v4 != nil {
		return v4
	}
	return ip
}

func ipToInt(ip net.IP) (*big.Int, int) {
	val := &big.Int{}
	val.SetBytes([]byte(ip))
	if len(ip) == net.IPv4len {
		return val, 32
	} else if len(ip) == net.IPv6len {
		return val, 128
	} else {
		panic(fmt.Errorf("Unsupported address length %d", len(ip)))
	}
}

func intToIP(ipInt *big.Int, bits int) net.IP {
	ipBytes := ipInt.Bytes()
	ret := make([]byte, bits/8)
	// Pack our IP bytes into the end of the return array,
	// since big.Int.Bytes() removes front zero padding.
	for i := 1; i <= len(ipBytes); i++ {
		ret[len(ret)-i] = ipBytes[len(ipBytes)-i]
	}
	return net.IP(ret)
}

func insertNumIntoIP(ip net.IP, num int, prefixLen int) net.IP {
	ipInt, totalBits := ipToInt(ip)
	bigNum := big.NewInt(int64(num))
	bigNum.Lsh(bigNum, uint(totalBits-prefixLen))
	ipInt.Or(ipInt, bigNum)
	return intToIP(ipInt, totalBits)
}
