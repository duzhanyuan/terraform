package aws

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsCloudFormation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudFormationCreate,
		Read:   resourceAwsCloudFormationRead,
		Update: resourceAwsCloudFormationUpdate,
		Delete: resourceAwsCloudFormationDelete,

		// CF_TODO: Determine ForceNew and Computed for the below schemas
		Schema: map[string]*schema.Schema{
			"capabilities": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				// ForceNew: true,
				// Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set: func(v interface{}) int {
					return hashcode.String(v.(string))
				},
			},
			"disable_rollback": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				// ForceNew: true,
				// Computed: true,
				Default:  false,
			},
			"notification_arns": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				// ForceNew: true,
				// Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set: func(v interface{}) int {
					return hashcode.String(v.(string))
				},
			},
			"on_failure": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				// ForceNew: true,
				// Computed: true,
				Default:  "ROLLBACK",
			},
			"parameters": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				// ForceNew: true,
				// Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"value": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"use_previous_value": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
				Set: func(v interface{}) int {
					return hashcode.String(v.(string))
				},
			},
			"stack_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				// ForceNew: true,
				// Computed: true,
			},
			"stack_policy_body": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				// ForceNew: true,
				// Computed: true,
			},
			"stack_policy_url": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				// ForceNew: true,
				// Computed: true,
			},
			"tags": tagsSchema(),
			"template_body": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				// ForceNew: true,
				// Computed: true,
				StateFunc: func(v interface{}) string {
					switch v.(type) {
					case string:
						hash := sha1.Sum([]byte(v.(string)))
						return hex.EncodeToString(hash[:])
					default:
						return ""
					}
				},
			},
			"template_url": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				// ForceNew: true,
				// Computed: true,
			},
			"timeout_in_minutes": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				// ForceNew: true,
				// Computed: true,
			},
			"use_previous_template": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				// ForceNew: true,
				// Computed: true,
			},
		},
	}
}

func resourceAwsCloudFormationCreate(d *schema.ResourceData, meta interface{}) error {
	// CF_TODO: Should below be cloudformationconn or conn?
	conn := meta.(*AWSClient).cloudformationSDKconn

	// Required StackInput parameters
	stackInputOpts := &cloudformation.CreateStackInput{
		StackName:        aws.String(d.Get("stack_name").(string)),
	}

	// Optional StackInput parameters
	// CF_TODO: Should there be space between these ifs?
	if v := d.Get("capabilities"); v != nil {
		stackInputOpts.Capabilities = expandStringListSDK(d.Get("capabilities").(*schema.Set).List())
	}

	if v := d.Get("disable_rollback"); v != nil {
		stackInputOpts.DisableRollback = aws.Boolean(d.Get("disable_rollback").(bool))
	}

	if v := d.Get("notification_arns"); v != nil {
		stackInputOpts.NotificationARNs = expandStringListSDK(d.Get("notification_arns").(*schema.Set).List())
	}

	if v := d.Get("on_failure"); v != nil {
		stackInputOpts.OnFailure = aws.String(d.Get("on_failure").(string))
	}

	if v := d.Get("parameters"); v != nil {
		// Expand the "parameter" set to aws-sdk-go compat []*cloudformation.Parameter
		parameters, err := expandCloudFormationParametersSDK(d.Get("parameters").(*schema.Set).List())

		// CF_TODO: Should there be spacing between conditionals and above statements?
		if err != nil {
			return err
		}

		stackInputOpts.Parameters = parameters
	}

	if v := d.Get("stack_policy_body"); v != nil {
		stackInputOpts.StackPolicyBody = aws.String(d.Get("stack_policy_body").(string))
	}

	if v := d.Get("stack_policy_url"); v != nil {
		stackInputOpts.StackPolicyURL = aws.String(d.Get("stack_policy_url").(string))
	}

	if v := d.Get("tags"); v != nil {
		stackInputOpts.Tags = tagsFromMapCloudFormation(d.Get("tags").(map[string]interface{}))
	}

	if v := d.Get("template_body"); v != nil {
		stackInputOpts.TemplateBody = aws.String(d.Get("template_body").(string))
	}

	if v := d.Get("template_url"); v != nil {
		stackInputOpts.TemplateURL = aws.String(d.Get("template_url").(string))
	}

	if v := d.Get("timeout_in_minutes"); v != nil {
		stackInputOpts.TimeoutInMinutes = aws.Long(int64(d.Get("timeout_in_minutes").(int)))
	}

	resp, err := conn.CreateStack(stackInputOpts)

	if err != nil {
		return fmt.Errorf("Error creating Stack: %s", err)
	}

	d.SetId(*resp.StackID)

	log.Printf("[INFO] Created CloudFormation Stack ID: %s", d.Id())

	return resourceAwsCloudFormationRead(d, meta)
}

func resourceAwsCloudFormationRead(d *schema.ResourceData, meta interface{}) error {
	// conn := meta.(*AWSClient).cloudformationSDKconn

	// stacks := make([]*cloudformation.Stack, 0, 0)
	// stacks, err := getStacksFunc(conn, d.Id(), "", nil)()

	// if err != nil {
		// return err
	// }

	// if stacks == nil {
		// return fmt.Errorf("Unable to find stacks within: %#v", resp.CloudFormations)
		// d.SetId("")
		// return nil
	// }

	return nil
}

func resourceAwsCloudFormationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudformationSDKconn

	// CF_TODO: Should opts be explicitly named per function?
	// Required StackUpdate parameters
	stackUpdateOpts := &cloudformation.UpdateStackInput{
		StackName: aws.String(d.Id()),
	}

	// Optional StackUpdate parameters to include regardless
	if v := d.Get("stack_policy_during_update_body"); v != nil {
		stackUpdateOpts.StackPolicyDuringUpdateBody = aws.String(d.Get("stack_policy_during_update_body").(string))
	}

	if v := d.Get("stack_policy_during_update_url"); v != nil {
		stackUpdateOpts.StackPolicyDuringUpdateURL = aws.String(d.Get("stack_policy_during_update_url").(string))
	}

	// Optional StackUpdate parameters only to include if changed
	if d.HasChange("capabilities") {
		stackUpdateOpts.Capabilities = expandStringListSDK(d.Get("capabilities").(*schema.Set).List())
	}

	if d.HasChange("notification_arns") {
		stackUpdateOpts.NotificationARNs = expandStringListSDK(d.Get("notification_arns").(*schema.Set).List())
	}

	if d.HasChange("parameters") {
		// Expand the "parameter" set to aws-sdk-go compat []*cloudformation.Parameter
		parameters, err := expandCloudFormationParametersSDK(d.Get("parameters").(*schema.Set).List())

		// CF_TODO: Should there be spacing between conditionals and above statements?
		if err != nil {
			return err
		}

		stackUpdateOpts.Parameters = parameters
	}

	if d.HasChange("stack_policy_body") {
		stackUpdateOpts.StackPolicyBody = aws.String(d.Get("stack_policy_body").(string))
	}

	if d.HasChange("stack_policy_url") {
		stackUpdateOpts.StackPolicyURL = aws.String(d.Get("stack_policy_url").(string))
	}

	if  d.HasChange("template_body") {
		stackUpdateOpts.TemplateBody = aws.String(d.Get("template_body").(string))
	}

	if d.HasChange("template_url") {
		stackUpdateOpts.TemplateURL = aws.String(d.Get("template_url").(string))
	}

	log.Printf("[DEBUG] CloudFormation Stack update configuration: %#v", stackUpdateOpts)
	resp, err := conn.UpdateStack(stackUpdateOpts)

	if err != nil {
		d.Partial(true)
		return fmt.Errorf("Error updating CloudFormation Stack: %s", err)
	}

	log.Printf("[INFO] Updated CloudFormation Stack ID: %s", *resp.StackID)

	return resourceAwsCloudFormationRead(d, meta)
}


func resourceAwsCloudFormationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudformationSDKconn

	// Required StackUpdate parameters
	stackDeleteOpts := &cloudformation.DeleteStackInput{
		StackName: aws.String(d.Id()),
	}

	_, err := conn.DeleteStack(stackDeleteOpts)

	if err != nil {
		return fmt.Errorf("Error deleting CloudFormation Stack: %s", err)
	}

	log.Printf("[INFO] Deleted CloudFormation Stack ID: %s", d.Id())

	return nil
}

// getStacksFunc returns a resource.getStacksFunc that
// is used to get all CloudFormation Stacks
// func getStacksFunc(
	// conn *cloudformation.CloudFormation, id string, nextToken string,
	// stacks []*cloudformation.Stack) (cloudformation.Stack, error) {

	// req := &cloudformation.DescribeStacksInput{
		// StackName: aws.String(id),
	// }

	// if nextToken != nil {
		// req.NextToken = aws.String(nextToken)
	// }

	// resp, err := conn.DescribeStacks(req)

	// if err != nil {
		// return stacks, fmt.Errorf("Error retrieving DescribeStacks: %s", err)
	// }

	// if resp.NextToken != nil {
		// stacks := getStacksFunc(conn, id, resp.nextToken, stacks)
	// }

	// return stacks, nil
// }
