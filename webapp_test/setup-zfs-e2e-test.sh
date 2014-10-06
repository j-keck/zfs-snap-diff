#!/usr/bin/env bash
#
#
ZPOOL_NAME=stuff
ZFS_NAME="${ZPOOL_NAME}/zsd-e2e-tests"

# abort if ZFS_NAME is ZPOOL_NAME
if [ "${ZPOOL_NAME}" == "${ZFS_NAME}" ]; then
  echo "ZPOOL_NAME == ZFS_NAME - ABORT!"
  exit
fi

#
# destory test zfs
#
CMD="zfs destroy -r ${ZFS_NAME}"
echo destory old zfs: ${CMD}
${CMD} || exit


#
# create datasets
#
declare -a DATASETS=(${ZFS_NAME} "${ZFS_NAME}/child1" "${ZFS_NAME}/child2" "${ZFS_NAME}/child3")
for DATASET in "${DATASETS[@]}"
do
  CMD="zfs create ${DATASET}"
  echo create zfs: ${CMD}
  ${CMD} || exit 
done  


#
# create test files
#
echo create files...
for DATASET in "${DATASETS[@]}"
do
  cat << EOF > /${DATASET}/file1
first line
second line
thrid line
EOF
done


#
# create a snapshot
#
echo create snapshot
CMD="zfs snapshot -r ${ZFS_NAME}@snap1"
echo create snapshot: ${CMD}
${CMD} || exit


#
# update test files
#
echo update files
for DATASET in "${DATASETS[@]}"
do
  DATASET_NAME=${DATASET##*/}
  sed -i '' "s/second/SECOND:${DATASET_NAME}/" /${DATASET}/file1
done

