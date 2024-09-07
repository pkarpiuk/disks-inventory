# disks-inventory
Tool to document contents of your disks, for search, visualize, duplication detection etc.

list: dump list of detected disks (having label)

--disks: list of disks to process (by label, device name, or mountpoint)
--tmpdir: where to unpack all archives
--datadir: inventory directory
--threads: how many threads (default: as many as processed disks)