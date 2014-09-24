#!/usr/bin/env bash
#
#
ZPOOL_NAME=stuff
ZFS_NAME="${ZPOOL_NAME}/zsd-e2e-tests"
DIR="/${ZFS_NAME}"

CMD="zfs destroy -r ${ZFS_NAME}"
echo destory old zfs: ${CMD}
${CMD} || exit


CMD="zfs create ${ZFS_NAME}"
echo create zfs: ${CMD}
${CMD} || exit 


echo create files...
cat << EOF > ${DIR}/file1
first line
second line
thrid line
EOF


CMD="zfs snapshot ${ZFS_NAME}@snap1"
echo create snapshot: ${CMD}
${CMD} || exit


echo update file1
sed -i '' 's/second/SECOND/' ${DIR}/file1

