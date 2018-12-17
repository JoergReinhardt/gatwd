#!/bin/zsh
# takes a path to a folder containing json formatted kubernetes manifests and a
# path to an output dir as it's second argument. generates a single dir by
# concatenating the manifests and wrapping them into a single object, by giving
# each it's own field with the dirpath as it's key.
dir=${1:-${HOME}/.config/kubextract/cluster-backup}
output=${2:-${HOME}/.config/kubextract/cluster-dump.json}

cd ${dir}

echo '' >  ${output}
for man in $(find ${dir} -type f -name '*.json')
do
filename=$(basename ${man})
manifest=$(jq . ${man})
manlist=$(printf '{"%s><": %s\n}\n' ${filename}  ${manifest} >> ${output})
done
