package purestorage

import (
	"github.com/devans10/go-purestorage/flasharray"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourcePureHost() *schema.Resource {
	return &schema.Resource{
		Create: resourcePureHostCreate,
		Read:   resourcePureHostRead,
		Update: resourcePureHostUpdate,
		Delete: resourcePureHostDelete,
		Importer: &schema.ResourceImporter{
			State: resourcePureHostImport,
		},
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Name of the host",
				Required:    true,
			},
			"iqn": &schema.Schema{
				Type:        schema.TypeList,
				Description: "List of iSCSI qualified names (IQNs) to the specified host.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: false,
				Optional: true,
			},
			"wwn": &schema.Schema{
				Type:        schema.TypeList,
				Description: "List of Fibre Channel worldwide names (WWNs) to the specified host.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: false,
				Optional: true,
			},
			"nqn": &schema.Schema{
				Type:        schema.TypeList,
				Description: "List of NVMeF qualified names (NQNs) to the specified host.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: false,
				Optional: true,
			},
			"host_password": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Host password for CHAP authentication.",
				Computed:    true,
				Optional:    true,
			},
			"host_user": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Host username for CHAP authentication.",
				Optional:    true,
				Default:     "",
			},
			"personality": &schema.Schema{
				Type:         schema.TypeString,
				Description:  "Determines how the Purity system tunes the protocol used between the array and the initiator.",
				Optional:     true,
				Default:      "",
				ValidateFunc: validation.StringInSlice([]string{"", "aix", "esxi", "hitachi-vsp", "hpux", "oracle-vm-server", "solaris", "vms"}, false),
			},
			"preferred_array": &schema.Schema{
				Type:        schema.TypeList,
				Description: "List of preferred arrays.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				Default:  nil,
			},
			"hgroup": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"target_password": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Target password for CHAP authentication.",
				Computed:    true,
				Optional:    true,
			},
			"target_user": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Target username for CHAP authentication.",
				Optional:    true,
				Default:     "",
			},
			"connected_volumes": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				Default:  nil,
			},
		},
	}
}

func resourcePureHostCreate(d *schema.ResourceData, m interface{}) error {

	d.Partial(true)
	client := m.(*flasharray.Client)
	var h *flasharray.Host
	var err error

	v, _ := d.GetOk("name")

	data := make(map[string]interface{})

	if wl, ok := d.GetOk("wwn"); ok {
		var wwnlist []string
		for _, element := range wl.([]interface{}) {
			wwnlist = append(wwnlist, element.(string))
		}
		data["wwnlist"] = wwnlist
	}

	if il, ok := d.GetOk("iqn"); ok {
		var iqnlist []string
		for _, element := range il.([]interface{}) {
			iqnlist = append(iqnlist, element.(string))
		}
		data["iqnlist"] = iqnlist
	}

	if nl, ok := d.GetOk("nqn"); ok {
		var nqnlist []string
		for _, element := range nl.([]interface{}) {
			nqnlist = append(nqnlist, element.(string))
		}
		data["nqnlist"] = nqnlist
	}

	if pa, ok := d.GetOk("preferred_array"); ok {
		var preferred_array []string
		for _, element := range pa.([]interface{}) {
			preferred_array = append(preferred_array, element.(string))
		}
		data["preferred_array"] = preferred_array
	}

	if len(data) > 0 {
		h, err = client.Hosts.CreateHost(v.(string), data)
		if err != nil {
			return err
		}
	} else {
		h, err = client.Hosts.CreateHost(v.(string), nil)
		if err != nil {
			return err
		}
	}
	d.SetPartial("name")
	d.SetPartial("wwn")
	d.SetPartial("iqn")
	d.SetPartial("nqn")
	d.SetPartial("preferred_array")

	chap_details := make(map[string]interface{})
	if host_password, ok := d.GetOk("host_password"); ok {
		chap_details["host_password"] = host_password.(string)
	}

	if host_user, ok := d.GetOk("host_user"); ok {
		chap_details["host_user"] = host_user.(string)
	}

	if target_password, ok := d.GetOk("target_password"); ok {
		chap_details["target_password"] = target_password.(string)
	}

	if target_user, ok := d.GetOk("target_user"); ok {
		chap_details["target_user"] = target_user.(string)
	}

	if len(chap_details) > 0 {
		h, err = client.Hosts.SetHost(h.Name, chap_details)
		if err != nil {
			return err
		}
	}
	d.SetPartial("host_password")
	d.SetPartial("host_user")
	d.SetPartial("target_password")
	d.SetPartial("target_user")

	if personality, ok := d.GetOk("personality"); ok {
		h, err = client.Hosts.SetHost(h.Name, map[string]string{"personality": personality.(string)})
		if err != nil {
			return err
		}
	}
	d.SetPartial("personality")

	var connected_volumes []string
	if cv, ok := d.GetOk("connected_volumes"); ok {
		for _, element := range cv.([]interface{}) {
			connected_volumes = append(connected_volumes, element.(string))
		}
	}

	if connected_volumes != nil {
		for _, volume := range connected_volumes {
			_, err = client.Hosts.ConnectHost(h.Name, volume)
			if err != nil {
				return err
			}
		}
	}

	d.Partial(false)

	d.SetId(h.Name)
	return resourcePureHostRead(d, m)
}

func resourcePureHostRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*flasharray.Client)

	host, _ := client.Hosts.GetHost(d.Id(), nil)

	if host == nil {
		d.SetId("")
		return nil
	}

	var connected_volumes []string
	cv, _ := client.Hosts.ListHostConnections(host.Name)
	for _, volume := range cv {
		connected_volumes = append(connected_volumes, volume.Vol)
	}

	d.Set("name", host.Name)
	d.Set("iqn", host.Iqn)
	d.Set("wwn", host.Wwn)
	d.Set("nqn", host.Nqn)
	d.Set("connected_volumes", connected_volumes)

	host, _ = client.Hosts.GetHost(d.Id(), map[string]string{"preferred_array": "true"})
	d.Set("preferred_array", host.PreferredArray)

	host, _ = client.Hosts.GetHost(d.Id(), map[string]string{"personality": "true"})
	d.Set("personality", host.Personality)

	host, _ = client.Hosts.GetHost(d.Id(), map[string]string{"chap": "true"})
	d.Set("host_password", host.HostPassword)
	d.Set("host_user", host.HostUser)
	d.Set("target_password", host.TargetPassword)
	d.Set("target_user", host.TargetUser)

	return nil
}

func resourcePureHostUpdate(d *schema.ResourceData, m interface{}) error {
	d.Partial(true)
	client := m.(*flasharray.Client)
	var h *flasharray.Host
	var err error

	if d.HasChange("name") {
		if h, err = client.Hosts.RenameHost(d.Id(), d.Get("name").(string)); err != nil {
			return err
		}
		d.SetId(h.Name)
	}
	d.SetPartial("name")

	if d.HasChange("wwn") {
		var wwnlist []string
		wl, _ := d.GetOk("wwn")
		for _, element := range wl.([]interface{}) {
			wwnlist = append(wwnlist, element.(string))
		}
		data := map[string]interface{}{"wwnlist": wwnlist}
		if _, err = client.Hosts.SetHost(d.Id(), data); err != nil {
			return err
		}
	}
	d.SetPartial("wwn")

	if d.HasChange("iqn") {
		var iqnlist []string
		il, _ := d.GetOk("iqn")
		for _, element := range il.([]interface{}) {
			iqnlist = append(iqnlist, element.(string))
		}
		data := map[string]interface{}{"iqnlist": iqnlist}
		if _, err = client.Hosts.SetHost(d.Id(), data); err != nil {
			return err
		}
	}
	d.SetPartial("iqn")

	if d.HasChange("nqn") {
		var nqnlist []string
		nl, _ := d.GetOk("nqn")
		for _, element := range nl.([]interface{}) {
			nqnlist = append(nqnlist, element.(string))
		}
		data := map[string]interface{}{"nqnlist": nqnlist}
		if _, err = client.Hosts.SetHost(d.Id(), data); err != nil {
			return err
		}
	}
	d.SetPartial("nqn")

	if d.HasChange("preferred_array") {
		var preferred_array []string
		pa, _ := d.GetOk("preferred_array")
		for _, element := range pa.([]interface{}) {
			preferred_array = append(preferred_array, element.(string))
		}
		data := map[string]interface{}{"preferred_array": preferred_array}
		if _, err = client.Hosts.SetHost(d.Id(), data); err != nil {
			return err
		}
	}
	d.SetPartial("preferred_array")

	chap_details := make(map[string]interface{})

	if d.HasChange("host_password") {
		chap_details["host_password"] = d.Get("host_password").(string)
	}

	if d.HasChange("host_user") {
		chap_details["host_user"] = d.Get("host_user").(string)
	}

	if d.HasChange("target_password") {
		chap_details["target_password"] = d.Get("target_password").(string)
	}

	if d.HasChange("target_user") {
		chap_details["target_user"] = d.Get("target_user").(string)
	}

	if len(chap_details) > 0 {
		if _, err = client.Hosts.SetHost(d.Id(), chap_details); err != nil {
			return err
		}
	}
	d.SetPartial("host_password")
	d.SetPartial("host_user")
	d.SetPartial("target_password")
	d.SetPartial("target_user")

	if d.HasChange("personality") {
		if _, err = client.Hosts.SetHost(d.Id(), map[string]string{"personality": d.Get("personality").(string)}); err != nil {
			return err
		}
	}
	d.SetPartial("personality")

	if d.HasChange("connected_volumes") {
		var connected_volumes []string
		cv, _ := d.GetOk("connected_volumes")
		for _, element := range cv.([]interface{}) {
			connected_volumes = append(connected_volumes, element.(string))
		}
		var current_volumes []string
		curvols, _ := client.Hosts.ListHostConnections(d.Id())
		for _, volume := range curvols {
			current_volumes = append(current_volumes, volume.Vol)
		}

		connect_volumes := difference(connected_volumes, current_volumes)
		for _, volume := range connect_volumes {
			if _, err = client.Hosts.ConnectHost(d.Id(), volume); err != nil {
				return err
			}
		}

		disconnect_volumes := difference(current_volumes, connected_volumes)
		for _, volume := range disconnect_volumes {
			if _, err = client.Hosts.DisconnectHost(d.Id(), volume); err != nil {
				return err
			}
		}
	}
	d.Partial(false)

	return resourcePureHostRead(d, m)
}

func resourcePureHostDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*flasharray.Client)

	var connected_volumes []string
	if cv, ok := d.GetOk("connected_volumes"); ok {
		for _, element := range cv.([]interface{}) {
			connected_volumes = append(connected_volumes, element.(string))
		}
	}
	if connected_volumes != nil {
		for _, volume := range connected_volumes {
			_, err := client.Hosts.DisconnectHost(d.Id(), volume)
			if err != nil {
				return err
			}
		}
	}

	_, err := client.Hosts.DeleteHost(d.Id())

	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func resourcePureHostImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*flasharray.Client)

	host, err := client.Hosts.GetHost(d.Id(), nil)

	if err != nil {
		return nil, err
	}

	var connected_volumes []string
	cv, _ := client.Hosts.ListHostConnections(host.Name)
	for _, volume := range cv {
		connected_volumes = append(connected_volumes, volume.Vol)
	}

	d.Set("name", host.Name)
	d.Set("iqn", host.Iqn)
	d.Set("wwn", host.Wwn)
	d.Set("nqn", host.Nqn)
	d.Set("connected_volumes", connected_volumes)

	host, _ = client.Hosts.GetHost(d.Id(), map[string]string{"preferred_array": "true"})
	d.Set("preferred_array", host.PreferredArray)

	host, _ = client.Hosts.GetHost(d.Id(), map[string]string{"personality": "true"})
	d.Set("personality", host.Personality)

	host, _ = client.Hosts.GetHost(d.Id(), map[string]string{"chap": "true"})
	d.Set("host_password", host.HostPassword)
	d.Set("host_user", host.HostUser)
	d.Set("target_password", host.TargetPassword)
	d.Set("target_user", host.TargetUser)

	return []*schema.ResourceData{d}, nil
}
