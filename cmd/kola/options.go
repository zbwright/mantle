// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/coreos/mantle/auth"
	"github.com/coreos/mantle/kola"
	"github.com/coreos/mantle/sdk"
)

var (
	outputDir          string
	kolaPlatform       string
	defaultTargetBoard = sdk.DefaultBoard()
	kolaPlatforms      = []string{"aws", "esx", "gce", "oci", "packet", "qemu"}
	kolaDefaultImages  = map[string]string{
		"amd64-usr": sdk.BuildRoot() + "/images/amd64-usr/latest/coreos_production_image.bin",
		"arm64-usr": sdk.BuildRoot() + "/images/arm64-usr/latest/coreos_production_image.bin",
	}

	kolaDefaultBIOS = map[string]string{
		"amd64-usr": "bios-256k.bin",
		"arm64-usr": sdk.BuildRoot() + "/images/arm64-usr/latest/coreos_production_qemu_uefi_efi_code.fd",
	}
)

func init() {
	sv := root.PersistentFlags().StringVar
	bv := root.PersistentFlags().BoolVar

	// general options
	sv(&outputDir, "output-dir", "", "Temporary output directory for test data and logs")
	sv(&kola.TorcxManifestFile, "torcx-manifest", "", "Path to a torcx manifest that should be made available to tests")
	root.PersistentFlags().StringVarP(&kolaPlatform, "platform", "p", "qemu", "VM platform: "+strings.Join(kolaPlatforms, ", "))
	root.PersistentFlags().IntVarP(&kola.TestParallelism, "parallel", "j", 1, "number of tests to run in parallel")
	sv(&kola.TAPFile, "tapfile", "", "file to write TAP results to")
	sv(&kola.Options.BaseName, "basename", "kola", "Cluster name prefix")

	// QEMU-specific options
	sv(&kola.QEMUOptions.Board, "board", defaultTargetBoard, "target board")
	sv(&kola.QEMUOptions.DiskImage, "qemu-image", "", "path to CoreOS disk image")
	sv(&kola.QEMUOptions.BIOSImage, "qemu-bios", "", "BIOS to use for QEMU vm")

	// gce-specific options
	sv(&kola.GCEOptions.Image, "gce-image", "projects/coreos-cloud/global/images/family/coreos-alpha", "GCE image, full api endpoints names are accepted if resource is in a different project")
	sv(&kola.GCEOptions.Project, "gce-project", "coreos-gce-testing", "GCE project name")
	sv(&kola.GCEOptions.Zone, "gce-zone", "us-central1-a", "GCE zone name")
	sv(&kola.GCEOptions.MachineType, "gce-machinetype", "n1-standard-1", "GCE machine type")
	sv(&kola.GCEOptions.DiskType, "gce-disktype", "pd-ssd", "GCE disk type")
	sv(&kola.GCEOptions.Network, "gce-network", "default", "GCE network")
	bv(&kola.GCEOptions.ServiceAuth, "gce-service-auth", false, "for non-interactive auth when running within GCE")
	sv(&kola.GCEOptions.JSONKeyFile, "gce-json-key", "", "use a service account's JSON key for authentication")

	// aws-specific options
	defaultRegion := os.Getenv("AWS_REGION")
	if defaultRegion == "" {
		defaultRegion = "us-west-2"
	}
	// Container Linux 1430.0.0 (alpha) on us-west-2
	sv(&kola.AWSOptions.Region, "aws-region", defaultRegion, "AWS region")
	sv(&kola.AWSOptions.AMI, "aws-ami", "alpha", `AWS AMI ID, or (alpha|beta|stable) to use the latest image`)
	sv(&kola.AWSOptions.InstanceType, "aws-type", "t2.small", "AWS instance type")
	sv(&kola.AWSOptions.SecurityGroup, "aws-sg", "kola", "AWS security group name")

	// packet-specific options
	sv(&kola.PacketOptions.ConfigPath, "packet-config-file", "", "Packet config file (default \"~/"+auth.PacketConfigPath+"\")")
	sv(&kola.PacketOptions.Profile, "packet-profile", "", "Packet profile (default \"default\")")
	sv(&kola.PacketOptions.ApiKey, "packet-api-key", "", "Packet API key (overrides config file)")
	sv(&kola.PacketOptions.Project, "packet-project", "", "Packet project UUID (overrides config file)")
	sv(&kola.PacketOptions.Facility, "packet-facility", "sjc1", "Packet facility code")
	sv(&kola.PacketOptions.Plan, "packet-plan", "", "Packet plan slug (default board-dependent, e.g. \"baremetal_0\")")
	sv(&kola.PacketOptions.InstallerImageBaseURL, "packet-installer-image-base-url", "", "Packet installer image base URL, non-https (default board-dependent, e.g. \"http://stable.release.core-os.net/amd64-usr/current\")")
	sv(&kola.PacketOptions.ImageURL, "packet-image-url", "", "Packet image URL (default board-dependent, e.g. \"https://alpha.release.core-os.net/amd64-usr/current/coreos_production_packet_image.bin.bz2\")")
	sv(&kola.PacketOptions.StorageURL, "packet-storage-url", "gs://users.developer.core-os.net/"+os.Getenv("USER")+"/mantle", "Google Storage base URL for temporary uploads")

	// esx-specific options
	sv(&kola.ESXOptions.Server, "esx-server", "", "ESX server")
	sv(&kola.ESXOptions.Profile, "esx-profile", "", "ESX profile (default \"default\")")
	sv(&kola.ESXOptions.BaseVMName, "esx-base-vm", "", "ESX base VM name")

	// oci-specific options
	sv(&kola.OCIOptions.Region, "oci-region", "", "OCI region")
	sv(&kola.OCIOptions.Image, "oci-image", "", "OCI image id")
	sv(&kola.OCIOptions.Shape, "oci-shape", "VM.Standard1.1", "OCI shape")
}

// Sync up the command line options if there is dependency
func syncOptions() error {
	kola.PacketOptions.Board = kola.QEMUOptions.Board
	kola.PacketOptions.GSOptions = &kola.GCEOptions

	ok := false
	for _, platform := range kolaPlatforms {
		if platform == kolaPlatform {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("unsupport platform %q", kolaPlatform)
	}

	image, ok := kolaDefaultImages[kola.QEMUOptions.Board]
	if !ok {
		return fmt.Errorf("unsupport board %q", kola.QEMUOptions.Board)
	}

	if kola.QEMUOptions.DiskImage == "" {
		kola.QEMUOptions.DiskImage = image
	}

	if kola.QEMUOptions.BIOSImage == "" {
		kola.QEMUOptions.BIOSImage = kolaDefaultBIOS[kola.QEMUOptions.Board]
	}

	return nil
}
