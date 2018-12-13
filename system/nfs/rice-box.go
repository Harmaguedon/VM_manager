package nfs

import (
	"time"

	"github.com/GeertJohan/go.rice/embedded"
)

func init() {

	// define files
	file2 := &embedded.EmbeddedFile{
		Filename:    "block_device_mount.sh",
		FileModTime: time.Unix(1542618620, 0),
		Content:     string("#!/usr/bin/env bash\n#\n# Copyright 2018, CS Systemes d'Information, http://www.c-s.fr\n#\n# Licensed under the Apache License, Version 2.0 (the \"License\");\n# you may not use this file except in compliance with the License.\n# You may obtain a copy of the License at\n#\n#     http://www.apache.org/licenses/LICENSE-2.0\n#\n# Unless required by applicable law or agreed to in writing, software\n# distributed under the License is distributed on an \"AS IS\" BASIS,\n# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.\n# See the License for the specific language governing permissions and\n# limitations under the License.\n#\n# block_device_mount.sh\n# Creates a filesystem on a device and mounts it\n\nset -u -o pipefail\n\nfunction print_error {\n    read line file <<<$(caller)\n    echo \"An error occurred in line $line of file $file:\" \"{\"`sed \"${line}q;d\" \"$file\"`\"}\" >&2\n}\ntrap print_error ERR\n\nfunction dns_fallback {\n    grep nameserver /etc/resolv.conf && return 0\n    echo -e \"nameserver 1.1.1.1\\n\" > /tmp/resolv.conf\n    sudo cp /tmp/resolv.conf /etc/resolv.conf\n    return 0\n}\n\ndns_fallback\n\nmkfs -F -t {{.FileSystem}} \"{{.Device}}\" && \\\nmkdir -p \"{{.MountPoint}}\" && \\\necho \"{{.Device}} {{.MountPoint}} {{.FileSystem}} defaults 0 2\" >>/etc/fstab && \\\nmount \"{{.MountPoint}}\" && \\\nchmod a+rwx \"{{.MountPoint}}\"\n"),
	}
	file3 := &embedded.EmbeddedFile{
		Filename:    "block_device_unmount.sh",
		FileModTime: time.Unix(1542618620, 0),
		Content:     string("#!/usr/bin/env bash\n#\n# Copyright 2018, CS Systemes d'Information, http://www.c-s.fr\n#\n# Licensed under the Apache License, Version 2.0 (the \"License\");\n# you may not use this file except in compliance with the License.\n# You may obtain a copy of the License at\n#\n#     http://www.apache.org/licenses/LICENSE-2.0\n#\n# Unless required by applicable law or agreed to in writing, software\n# distributed under the License is distributed on an \"AS IS\" BASIS,\n# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.\n# See the License for the specific language governing permissions and\n# limitations under the License.\n#\n# block_device_unmount.sh\n# Unmount a block device and removes the corresponding entry from /etc/fstab\n\nset -u -o pipefail\n\nfunction print_error {\n    read line file <<<$(caller)\n    echo \"An error occurred in line $line of file $file:\" \"{\"`sed \"${line}q;d\" \"$file\"`\"}\" >&2\n}\ntrap print_error ERR\n\nfunction dns_fallback {\n    grep nameserver /etc/resolv.conf && return 0\n    echo -e \"nameserver 1.1.1.1\\n\" > /tmp/resolv.conf\n    sudo cp /tmp/resolv.conf /etc/resolv.conf\n    return 0\n}\n\ndns_fallback\n\numount -l -f {{.Device}} && \\\nsed -i '\\:^{{.Device}}:d' /etc/fstab\n"),
	}
	file4 := &embedded.EmbeddedFile{
		Filename:    "nfs_client_install.sh",
		FileModTime: time.Unix(1542618620, 0),
		Content:     string("#!/usr/bin/env bash\n#\n# Copyright 2018, CS Systemes d'Information, http://www.c-s.fr\n#\n# Licensed under the Apache License, Version 2.0 (the \"License\");\n# you may not use this file except in compliance with the License.\n# You may obtain a copy of the License at\n#\n#     http://www.apache.org/licenses/LICENSE-2.0\n#\n# Unless required by applicable law or agreed to in writing, software\n# distributed under the License is distributed on an \"AS IS\" BASIS,\n# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.\n# See the License for the specific language governing permissions and\n# limitations under the License.\n#\n# Installs and configures\n\nset -u -o pipefail\n\nfunction print_error {\n    read line file <<<$(caller)\n    echo \"An error occurred in line $line of file $file:\" \"{\"`sed \"${line}q;d\" \"$file\"`\"}\" >&2\n}\ntrap print_error ERR\n\nfunction dns_fallback {\n    grep nameserver /etc/resolv.conf && return 0\n    echo -e \"nameserver 1.1.1.1\\n\" > /tmp/resolv.conf\n    sudo cp /tmp/resolv.conf /etc/resolv.conf\n    return 0\n}\n\ndns_fallback\n\n{{.reserved_BashLibrary}}\n\necho \"Install NFS client\"\ncase $LINUX_KIND in\n    debian|ubuntu)\n        export DEBIAN_FRONTEND=noninteractive\n        touch /var/log/lastlog\n        chgrp utmp /var/log/lastlog\n        chmod 664 /var/log/lastlog\n\n        sfRetry 3m 5 \"sfWaitForApt && apt -y update\"\n        sfRetry 5m 5 \"sfWaitForApt && apt-get install -qqy nfs-common\"\n        ;;\n\n    rhel|centos)\n        yum makecache fast\n        yum install -y nfs-utils\n        ;;\n\n    *)\n        echo \"Unsupported OS flavor '$LINUX_KIND'!\"\n        exit 1\nesac\n"),
	}
	file5 := &embedded.EmbeddedFile{
		Filename:    "nfs_client_share_mount.sh",
		FileModTime: time.Unix(1542618620, 0),
		Content:     string("#!/usr/bin/env bash\n#\n# Copyright 2018, CS Systemes d'Information, http://www.c-s.fr\n#\n# Licensed under the Apache License, Version 2.0 (the \"License\");\n# you may not use this file except in compliance with the License.\n# You may obtain a copy of the License at\n#\n#     http://www.apache.org/licenses/LICENSE-2.0\n#\n# Unless required by applicable law or agreed to in writing, software\n# distributed under the License is distributed on an \"AS IS\" BASIS,\n# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.\n# See the License for the specific language governing permissions and\n# limitations under the License.\n#\n# nfs_client_share_mount.sh\n#\n# Declares a remote share mount and mount it\n\nset -u -o pipefail\n\nfunction print_error {\n    read line file <<<$(caller)\n    echo \"An error occurred in line $line of file $file:\" \"{\"`sed \"${line}q;d\" \"$file\"`\"}\" >&2\n}\ntrap print_error ERR\n\nfunction dns_fallback {\n    grep nameserver /etc/resolv.conf && return 0\n    echo -e \"nameserver 1.1.1.1\\n\" > /tmp/resolv.conf\n    sudo cp /tmp/resolv.conf /etc/resolv.conf\n    return 0\n}\n\ndns_fallback\n\nmkdir -p \"{{.MountPoint}}\" && \\\nmount -o noac \"{{.Host}}:{{.Share}}\" \"{{.MountPoint}}\" && \\\necho \"{{.Host}}:{{.Share}} {{.MountPoint}}   nfs defaults,user,auto,noatime,intr,noac 0   0\" >>/etc/fstab\n"),
	}
	file6 := &embedded.EmbeddedFile{
		Filename:    "nfs_client_share_unmount.sh",
		FileModTime: time.Unix(1542618620, 0),
		Content:     string("#!/usr/bin/env bash\n#\n# Copyright 2018, CS Systemes d'Information, http://www.c-s.fr\n#\n# Licensed under the Apache License, Version 2.0 (the \"License\");\n# you may not use this file except in compliance with the License.\n# You may obtain a copy of the License at\n#\n#     http://www.apache.org/licenses/LICENSE-2.0\n#\n# Unless required by applicable law or agreed to in writing, software\n# distributed under the License is distributed on an \"AS IS\" BASIS,\n# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.\n# See the License for the specific language governing permissions and\n# limitations under the License.\n#\n# nfs_client_share_unmount.sh\n#\n# Unconfigures and unmounts a remote access to a NFS share\n\nset -u -o pipefail\n\nfunction print_error {\n    read line file <<<$(caller)\n    echo \"An error occurred in line $line of file $file:\" \"{\"`sed \"${line}q;d\" \"$file\"`\"}\" >&2\n}\ntrap print_error ERR\n\nfunction dns_fallback {\n    grep nameserver /etc/resolv.conf && return 0\n    echo -e \"nameserver 1.1.1.1\\n\" > /tmp/resolv.conf\n    sudo cp /tmp/resolv.conf /etc/resolv.conf\n    return 0\n}\n\ndns_fallback\n\numount -fl {{.Host}}:{{.Share}}\nsed -i '\\#^{{.Host}}:{{.Share}}#d' /etc/fstab\n"),
	}
	file7 := &embedded.EmbeddedFile{
		Filename:    "nfs_server_install.sh",
		FileModTime: time.Unix(1542618620, 0),
		Content:     string("#!/usr/bin/env bash\n#\n# Copyright 2018, CS Systemes d'Information, http://www.c-s.fr\n#\n# Licensed under the Apache License, Version 2.0 (the \"License\");\n# you may not use this file except in compliance with the License.\n# You may obtain a copy of the License at\n#\n#     http://www.apache.org/licenses/LICENSE-2.0\n#\n# Unless required by applicable law or agreed to in writing, software\n# distributed under the License is distributed on an \"AS IS\" BASIS,\n# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.\n# See the License for the specific language governing permissions and\n# limitations under the License.\n#\n# nfs_server_install.sh\n#\n# Installs and configures a NFS Server service\n\nset -u -o pipefail\n\nfunction print_error {\n    read line file <<<$(caller)\n    echo \"An error occurred in line $line of file $file:\" \"{\"`sed \"${line}q;d\" \"$file\"`\"}\" >&2\n}\ntrap print_error ERR\n\nfunction dns_fallback {\n    grep nameserver /etc/resolv.conf && return 0\n    echo -e \"nameserver 1.1.1.1\\n\" > /tmp/resolv.conf\n    sudo cp /tmp/resolv.conf /etc/resolv.conf\n    return 0\n}\n\ndns_fallback\n\n{{.reserved_BashLibrary}}\n\necho \"Install NFS server\"\n\ncase $LINUX_KIND in\n    debian|ubuntu)\n        export DEBIAN_FRONTEND=noninteractive\n        touch /var/log/lastlog\n        chgrp utmp /var/log/lastlog\n        chmod 664 /var/log/lastlog\n        sfWaitForApt && apt-get update && sfWaitForApt && apt-get install -qqy nfs-common nfs-kernel-server\n        ;;\n\n    rhel|centos)\n        yum makecache fast\n        yum install -y nfs-utils\n        systemctl enable rpcbind\n        systemctl enable nfs-server\n        systemctl enable nfs-lock\n        systemctl enable nfs-idmap\n        systemctl start rpcbind\n        systemctl start nfs-server\n        systemctl start nfs-lock\n        systemctl start nfs-idmap\n        ;;\n\n    *)\n        echo \"Unsupported operating system '$LINUX_KIND'\"\n        exit 1\n        ;;\nesac\n"),
	}
	file8 := &embedded.EmbeddedFile{
		Filename:    "nfs_server_path_export.sh",
		FileModTime: time.Unix(1542618620, 0),
		Content:     string("#!/usr/bin/env bash\n#\n# Copyright 2018, CS Systemes d'Information, http://www.c-s.fr\n#\n# Licensed under the Apache License, Version 2.0 (the \"License\");\n# you may not use this file except in compliance with the License.\n# You may obtain a copy of the License at\n#\n#     http://www.apache.org/licenses/LICENSE-2.0\n#\n# Unless required by applicable law or agreed to in writing, software\n# distributed under the License is distributed on an \"AS IS\" BASIS,\n# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.\n# See the License for the specific language governing permissions and\n# limitations under the License.\n#\n# nfs_server_path_export.sh\n#\n# Configures the NFS export of a local path\n\nset -u -o pipefail\n\nfunction print_error {\n    read line file <<<$(caller)\n    echo \"An error occurred in line $line of file $file:\" \"{\"`sed \"${line}q;d\" \"$file\"`\"}\" >&2\n}\ntrap print_error ERR\n\nfunction dns_fallback {\n    grep nameserver /etc/resolv.conf && return 0\n    echo -e \"nameserver 1.1.1.1\\n\" > /tmp/resolv.conf\n    sudo cp /tmp/resolv.conf /etc/resolv.conf\n    return 0\n}\n\ndns_fallback\n\n# Determines the FSID value to use\nFSIDs=$(cat /etc/exports | sed -r 's/ /\\n/g' | grep fsid= | sed -r 's/.+fsid=([[:alnum:]]+),.*/\\1/g' | uniq | sort -n)\nLAST_FSID=$(echo \"$FSIDs\" | tail -n 1)\nif [ -z \"$LAST_FSID\" ]; then\n    FSID=1\nelse\n    FSID=$((LAST_FSID + 1))\nfi\n\n# Adapts ACL\nACCESS_RIGHTS=\"{{.AccessRights}}\"\nFILTERED_ACCESS_RIGHTS=\nif [ -z \"$ACCESS_RIGHTS\" ]; then\n    # No access rights, using default ones\n    FILTERED_ACCESS_RIGHTS=\"*(rw,fsid=$FSID,sync,no_root_squash,no_subtree_check)\"\nelse\n    # Wants to ensure FSID is valid otherwise updates it\n    ACL=$(echo $ACCESS_RIGHTS | sed -r 's/\\((.*)\\)')\n    if [ ! -z \"$ACL\" ]; then\n        # If there is something between parenthesis, checks if there is some fsid directive, and check the values\n        # are not already used for other shares\n        ACL_FSIDs=$(echo $ACL | sed -r 's/ /\\n/g' | grep fsid= | sed -r 's/.+fsid=([[:alnum:]]+),.*/\\1/g' | uniq | sort -n)\n        for f in $ACL_FSIDs; do\n            echo $FSIDs | grep \"^${f}$\" && {\n                # FSID value is already used, updating the Access Rights to use the calculated new FSID\n                FILTERED_ACCESS_RIGHTS=$(echo $ACCESS_RIGHTS | sed -r 's/fsid=[[:numeric:]]*/fsid=$FSID/g')\n            }\n            break\n        done\n        if [ -z $FILTERED_ACCESS_RIGHTS ]; then\n            # No updated access rights, with something between parenthesis, adding fsid= directive\n            FILTERED_ACCESS_RIGHTS=$(echo $ACCESS_RIGHTS | sed -r 's/\\)/,fsid=$FSID\\)/g')\n        fi\n    else\n        # No updated access rights without anything between parenthesis, adding fsid= directive\n        FILTERED_ACCESS_RIGHTS=$(echo $ACCESS_RIGHTS | sed -r 's/\\)/fsid=$FSID/g')\n    fi\nfi\n#VPL: case not managed: nothing between braces...\n\n# Create exported dir if necessary\nmkdir -p \"{{.Path}}\"\nchmod a+rwx \"{{.Path}}\"\n\n# Configures export\necho \"{{.Path}} $FILTERED_ACCESS_RIGHTS\" >>/etc/exports\n\n# Updates exports\nexportfs -a\n"),
	}
	file9 := &embedded.EmbeddedFile{
		Filename:    "nfs_server_path_unexport.sh",
		FileModTime: time.Unix(1542618620, 0),
		Content:     string("#!/usr/bin/env bash\n#\n# Copyright 2018, CS Systemes d'Information, http://www.c-s.fr\n#\n# Licensed under the Apache License, Version 2.0 (the \"License\");\n# you may not use this file except in compliance with the License.\n# You may obtain a copy of the License at\n#\n#     http://www.apache.org/licenses/LICENSE-2.0\n#\n# Unless required by applicable law or agreed to in writing, software\n# distributed under the License is distributed on an \"AS IS\" BASIS,\n# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.\n# See the License for the specific language governing permissions and\n# limitations under the License.\n#\n# Unexports and unconfigures a NFS export of a local path\n\nset -u -o pipefail\n\nfunction print_error {\n    read line file <<<$(caller)\n    echo \"An error occurred in line $line of file $file:\" \"{\"`sed \"${line}q;d\" \"$file\"`\"}\" >&2\n}\ntrap print_error ERR\n\nfunction dns_fallback {\n    grep nameserver /etc/resolv.conf && return 0\n    echo -e \"nameserver 1.1.1.1\\n\" > /tmp/resolv.conf\n    sudo cp /tmp/resolv.conf /etc/resolv.conf\n    return 0\n}\n\ndns_fallback\n\nsed -i '\\#^{{.Path}} #d' /etc/exports\nexportfs -ar\n"),
	}

	// define dirs
	dir1 := &embedded.EmbeddedDir{
		Filename:   "",
		DirModTime: time.Unix(1542040863, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file2, // "block_device_mount.sh"
			file3, // "block_device_unmount.sh"
			file4, // "nfs_client_install.sh"
			file5, // "nfs_client_share_mount.sh"
			file6, // "nfs_client_share_unmount.sh"
			file7, // "nfs_server_install.sh"
			file8, // "nfs_server_path_export.sh"
			file9, // "nfs_server_path_unexport.sh"

		},
	}

	// link ChildDirs
	dir1.ChildDirs = []*embedded.EmbeddedDir{}

	// register embeddedBox
	embedded.RegisterEmbeddedBox(`../nfs/scripts`, &embedded.EmbeddedBox{
		Name: `../nfs/scripts`,
		Time: time.Unix(1542040863, 0),
		Dirs: map[string]*embedded.EmbeddedDir{
			"": dir1,
		},
		Files: map[string]*embedded.EmbeddedFile{
			"block_device_mount.sh":       file2,
			"block_device_unmount.sh":     file3,
			"nfs_client_install.sh":       file4,
			"nfs_client_share_mount.sh":   file5,
			"nfs_client_share_unmount.sh": file6,
			"nfs_server_install.sh":       file7,
			"nfs_server_path_export.sh":   file8,
			"nfs_server_path_unexport.sh": file9,
		},
	})
}
