package ct

import (
	"testing"

	r "github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

var testProviders = map[string]terraform.ResourceProvider{
	"ct": Provider(),
}

const prettyResource = `
data "ct_config" "example" {
  pretty_print = true
  content = <<EOT
---
storage:
  filesystems:
    - name: "rootfs"
      mount:
        device: "/dev/disk/by-label/ROOT"
        format: "ext4"

  files:
    - path: "/etc/motd"
      filesystem: "rootfs"
      mode: 0644
      contents:
        inline: |
          Hello World!
EOT
}
`

const prettyExpected = `{
  "ignition": {
    "version": "2.0.0",
    "config": {}
  },
  "storage": {
    "filesystems": [
      {
        "name": "rootfs",
        "mount": {
          "device": "/dev/disk/by-label/ROOT",
          "format": "ext4"
        }
      }
    ],
    "files": [
      {
        "filesystem": "rootfs",
        "path": "/etc/motd",
        "contents": {
          "source": "data:,Hello%20World!%0A",
          "verification": {}
        },
        "mode": 420,
        "user": {},
        "group": {}
      }
    ]
  },
  "systemd": {},
  "networkd": {},
  "passwd": {}
}`

const ec2Resource = `
data "ct_config" "ec2" {
  pretty_print = true
	platform = "ec2"
  content  = <<EOT
---
etcd:
  advertise_client_urls:       http://{PUBLIC_IPV4}:2379
  initial_advertise_peer_urls: http://{PRIVATE_IPV4}:2380
  listen_client_urls:          http://0.0.0.0:2379
  listen_peer_urls:            http://{PRIVATE_IPV4}:2380
  discovery:                   https://discovery.etcd.io/<token>
EOT
}
`

const ec2Expected = `{
  "ignition": {
    "version": "2.0.0",
    "config": {}
  },
  "storage": {},
  "systemd": {
    "units": [
      {
        "name": "etcd-member.service",
        "enable": true,
        "dropins": [
          {
            "name": "20-clct-etcd-member.conf",
            "contents": "[Unit]\nRequires=coreos-metadata.service\nAfter=coreos-metadata.service\n\n[Service]\nEnvironmentFile=/run/metadata/coreos\nExecStart=\nExecStart=/usr/lib/coreos/etcd-wrapper $ETCD_OPTS \\\n  --listen-peer-urls=\"http://${COREOS_EC2_IPV4_LOCAL}:2380\" \\\n  --listen-client-urls=\"http://0.0.0.0:2379\" \\\n  --initial-advertise-peer-urls=\"http://${COREOS_EC2_IPV4_LOCAL}:2380\" \\\n  --advertise-client-urls=\"http://${COREOS_EC2_IPV4_PUBLIC}:2379\" \\\n  --discovery=\"https://discovery.etcd.io/\u003ctoken\u003e\""
          }
        ]
      }
    ]
  },
  "networkd": {},
  "passwd": {}
}`

func TestRender(t *testing.T) {
	r.UnitTest(t, r.TestCase{
		Providers: testProviders,
		Steps: []r.TestStep{
			r.TestStep{
				Config: prettyResource,
				Check: r.ComposeTestCheckFunc(
					r.TestCheckResourceAttr("data.ct_config.example", "rendered", prettyExpected),
				),
			},
			r.TestStep{
				Config: ec2Resource,
				Check: r.ComposeTestCheckFunc(
					r.TestCheckResourceAttr("data.ct_config.ec2", "rendered", ec2Expected),
				),
			},
		},
	})
}
