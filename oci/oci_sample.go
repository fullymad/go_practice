package main

import (
	"context"
	"fmt"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/oracle/oci-go-sdk/example/helpers"
	"github.com/oracle/oci-go-sdk/identity"
)

// For simplicity, FatalIfError has been used on all errors

// Some hard-coded values as constants
const avail_domain string = "Nbnl:US-ASHBURN-AD-1"
const instance_name string = "instance-20180901-2142"

// Helper function
func getTenancyId() string {
	tenancy_id, err := common.DefaultConfigProvider().TenancyOCID()
	helpers.FatalIfError(err)

	return tenancy_id
} // end getTenancyId

// Given the instance display name, get instance OCID in root compartment
func getInstanceIdByName(ctx context.Context, name string) (string, error) {
	instances := getInstances(ctx)
	
	for _, instance := range instances {
		if (*(instance.DisplayName) == name) {
			return *(instance.Id), nil
		}
	}
	
	return "", fmt.Errorf("Cannot find instance with name %s", name)
} // end getInstanceIdByName

// Get instances in root compartment of tenancy
func getInstances(ctx context.Context) []core.Instance{	
	cl, err := core.NewComputeClientWithConfigurationProvider(
		common.DefaultConfigProvider())
	helpers.FatalIfError(err)

	request := core.ListInstancesRequest {
		CompartmentId: common.String(getTenancyId()),
	}
	response, err := cl.ListInstances(ctx, request)
	helpers.FatalIfError(err)
	
	for _, instance := range response.Items {
		fmt.Printf("Instance %s: %s\n", *(instance.DisplayName),
			*(instance.Id))
	}
	fmt.Printf("\n")
	
	return response.Items
} // end getInstances

// Get volume IDs of all attached volumes in tenancy
func getAttachedVolumes(ctx context.Context) []core.VolumeAttachment {
	var iscsi core.IScsiVolumeAttachment

	cl, err := core.NewComputeClientWithConfigurationProvider(
		common.DefaultConfigProvider())
	helpers.FatalIfError(err)

	instance_id, err := getInstanceIdByName(ctx, instance_name)
	helpers.FatalIfError(err)

	request := core.ListVolumeAttachmentsRequest{
		CompartmentId: common.String(getTenancyId()),
		InstanceId: common.String(instance_id),
	}
	response, err := cl.ListVolumeAttachments(ctx, request)
	helpers.FatalIfError(err)

	for i, vol_attached := range response.Items {
		fmt.Printf("Volume %d: %s\n", i + 1, *vol_attached.GetVolumeId())
		fmt.Printf("Attachment OCID: %s\n", *vol_attached.GetId())

		iscsi = vol_attached.(core.IScsiVolumeAttachment)
		fmt.Printf("Iqn: %s\n\n", *iscsi.Iqn)
	}

	return response.Items
} // end getVolumes

// Get the availability domains in the tenancy
func getDomains(ctx context.Context) []identity.AvailabilityDomain {
	cl, err := identity.NewIdentityClientWithConfigurationProvider(
		common.DefaultConfigProvider())
	helpers.FatalIfError(err)

	request := identity.ListAvailabilityDomainsRequest{
		CompartmentId: common.String(getTenancyId()),
	}
	response, err := cl.ListAvailabilityDomains(
		context.Background(), request)
	helpers.FatalIfError(err)
	
	for i, avail_domain := range response.Items {
		fmt.Printf("Available domain %d: %v\n", i + 1, avail_domain)
	}
	fmt.Printf("\n");

	return response.Items
} // end getDomains

// Create a block storage volume in the root compartment and attach to
// known instance
func createBlockVolume(ctx context.Context) {
	cl_bs, err := core.NewBlockstorageClientWithConfigurationProvider(
		common.DefaultConfigProvider())
	helpers.FatalIfError(err)

	// The OCID of the tenancy containing the compartment
	tenancy_id := getTenancyId();

	vol_details := core.CreateVolumeDetails{
		AvailabilityDomain: common.String(avail_domain),
		CompartmentId: common.String(tenancy_id),
		DisplayName: common.String("Madhu's block volume through API"),
		SizeInGBs: common.Int64(53),
	}

	request := core.CreateVolumeRequest{
		CreateVolumeDetails: vol_details,
	}
	response, err := cl_bs.CreateVolume(ctx, request)
	helpers.FatalIfError(err)

	vol_id := *(response.Volume.Id)
	fmt.Printf("Created volume with OCID: %s\n", vol_id)

	// Callback function to check if volume has been provisioned
	shouldRetryFunc := func(r common.OCIOperationResponse) bool {
		if converted, ok := r.Response.(core.GetVolumeResponse); ok {
			return converted.LifecycleState != core.VolumeLifecycleStateAvailable
		}
		return true
	}

	poll_request := core.GetVolumeRequest{
		VolumeId: common.String(vol_id),
		RequestMetadata: helpers.GetRequestMetadataWithCustomizedRetryPolicy(shouldRetryFunc),
	}

	_, err = cl_bs.GetVolume(ctx, poll_request)
	helpers.FatalIfError(err)
	fmt.Printf("Volume provisioned\n")

	// Attach the volume to the instance
	cl_c, err := core.NewComputeClientWithConfigurationProvider(
		common.DefaultConfigProvider())
	helpers.FatalIfError(err)
	
	vol_instance, err := getInstanceIdByName(ctx, instance_name)
	helpers.FatalIfError(err)

	iscsi_vol := core.AttachIScsiVolumeDetails {
		InstanceId: common.String(vol_instance),
		VolumeId: common.String(vol_id),
		DisplayName: common.String("Attached through API"),
	}

	attach_request := core.AttachVolumeRequest {
		AttachVolumeDetails: iscsi_vol,
	}
	attach_response, err := cl_c.AttachVolume(ctx, attach_request)
	helpers.FatalIfError(err)
	
	vol_attached := attach_response.VolumeAttachment
	
	fmt.Printf("Attached Iscsi volume: %s\n",
		*(vol_attached.GetDisplayName()))
	
	return
} // end createBlockVolume

func main() {
	ctx := context.Background()

	getInstances(ctx)
	getAttachedVolumes(ctx)
	getDomains(ctx)
	// Leave call to createBlockVolume commented until ready to create
	//createBlockVolume(ctx)

	return
} //end main
