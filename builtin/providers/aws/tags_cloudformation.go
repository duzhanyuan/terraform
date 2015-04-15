package aws

import (
    "fmt"
	"log"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform/helper/schema"
)

// CF_TODO: Should these files and functions be called CloudFormationSDK or just CloudFormation?
// CF_TODO: Should this function just be removed since it's not implemented yet?
// setTags is a helper to set the tags for a resource. It expects the
// tags field to be named "tags"
func setTagsCloudFormation(conn *cloudformation.CloudFormation, d *schema.ResourceData) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		create, remove := diffTagsCloudFormation(tagsFromMapCloudFormation(o), tagsFromMapCloudFormation(n))

		// Set tags
		if len(remove) > 0 {
			log.Printf("[DEBUG] Removing tags: %#v", remove)

            // _, err := conn.DeleteTags(&cloudformation.DeleteTagsInput{
                // Resources: []*string{aws.String(d.Id())},
                // Tags:      remove,
            // })

            // if err != nil {
                // return err
            // }
		}

		if len(create) > 0 {
			log.Printf("[DEBUG] Creating tags: %#v", create)

            // _, err := conn.CreateTags(&cloudformation.CreateTagsInput{
                // Resources: []*string{aws.String(d.Id())},
                // Tags:      create,
            // })

            // if err != nil {
                // return err
            // }
		}
	}

    errMsg := "[ERROR] setTagsCloudFormation is not implemented, the AWS CloudFormation API does not yet support it"
	log.Printf(errMsg)
    return fmt.Errorf(errMsg)
	// return nil

}

// diffTags takes our tags locally and the ones remotely and returns
// the set of tags that must be created, and the set of tags that must
// be destroyed.
func diffTagsCloudFormation(oldTags, newTags []*cloudformation.Tag) ([]*cloudformation.Tag, []*cloudformation.Tag) {
	// First, we're creating everything we have
	create := make(map[string]interface{})
	for _, t := range newTags {
		create[*t.Key] = *t.Value
	}

	// Build the list of what to remove
	var remove []*cloudformation.Tag
	for _, t := range oldTags {
		old, ok := create[*t.Key]
		if !ok || old != *t.Value {
			// Delete it!
			remove = append(remove, t)
		}
	}

	return tagsFromMapCloudFormation(create), remove
}

// tagsFromMap returns the tags for the given map of data.
func tagsFromMapCloudFormation(m map[string]interface{}) []*cloudformation.Tag {
	result := make([]*cloudformation.Tag, 0, len(m))
	for k, v := range m {
		result = append(result, &cloudformation.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		})
	}

	return result
}

// tagsToMap turns the list of tags into a map.
func tagsToMapCloudFormation(ts []*cloudformation.Tag) map[string]string {
	result := make(map[string]string)
	for _, t := range ts {
		result[*t.Key] = *t.Value
	}

	return result
}
