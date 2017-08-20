from collections import defaultdict

def getsubs(loc, s):
    substr = s[loc:]
    i = -1
    while(substr):
        yield substr
        substr = s[loc:i]
        i -= 1

def longestRepetitiveSubstring(r, minocc=3):
    occ = defaultdict(int)
    # tally all occurrences of all substrings
    for i in range(len(r)):
        for sub in getsubs(i,r):
            occ[sub] += 1

    # filter out all substrings with fewer than minocc occurrences
    occ_minocc = [k for k,v in occ.items() if v >= minocc]

    if occ_minocc:
        maxkey =  max(occ_minocc, key=len)
        return maxkey, occ[maxkey]
    else:
        raise ValueError("no repetitions of any substring of '%s' with %d or more occurrences" % (r,minocc))