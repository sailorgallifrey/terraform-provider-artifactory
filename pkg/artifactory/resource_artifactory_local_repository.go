package artifactory

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"net/http"
)

func resourceArtifactoryLocalRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceLocalRepositoryCreate,
		Read:   resourceLocalRepositoryRead,
		Update: resourceLocalRepositoryUpdate,
		Delete: resourceLocalRepositoryDelete,
		Exists: resourceLocalRepositoryExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"package_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: repoTypeValidator,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"notes": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"includes_pattern": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"excludes_pattern": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"repo_layout_ref": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"handle_releases": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"handle_snapshots": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"max_unique_snapshots": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"debian_trivial_layout": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"checksum_policy_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"max_unique_tags": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"snapshot_version_behavior": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"suppress_pom_consistency_checks": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"blacked_out": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"property_sets": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
			},
			"archive_browsing_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"calculate_yum_metadata": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"yum_root_depth": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"docker_api_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"enable_file_lists_indexing": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"xray_index": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"force_nuget_authentication": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

type MessyRepo struct {
	services.LocalRepositoryBaseParams
	services.CommonMavenGradleLocalRepositoryParams
	services.DebianLocalRepositoryParams
	services.DockerLocalRepositoryParams
	services.RpmLocalRepositoryParams
	ForceNugetAuthentication bool `json:"forceNugetAuthentication"`
}

func unmarshalLocalRepository(data *schema.ResourceData) MessyRepo {
	d := &ResourceData{ResourceData: data}
	repo := MessyRepo{}

	repo.Rclass = "local"
	repo.Key = d.getString("key", false)
	repo.PackageType = d.getString("package_type", false)
	repo.Description = d.getString("description", false)
	repo.Notes = d.getString("notes", false)
	repo.DebianTrivialLayout = d.getBoolRef("debian_trivial_layout", false)
	repo.IncludesPattern = d.getString("includes_pattern", false)
	repo.ExcludesPattern = d.getString("excludes_pattern", false)
	repo.RepoLayoutRef = d.getString("repo_layout_ref", false)
	repo.MaxUniqueTags = d.getInt("max_unique_tags", false)
	repo.BlackedOut = d.getBoolRef("blacked_out", false)
	repo.CalculateYumMetadata = d.getBoolRef("calculate_yum_metadata", false)
	repo.YumRootDepth = d.getInt("yum_root_depth", false)
	repo.ArchiveBrowsingEnabled = d.getBoolRef("archive_browsing_enabled", false)
	repo.DockerApiVersion = d.getString("docker_api_verision", false)
	repo.EnableFileListsIndexing = d.getBoolRef("enable_file_lists_indexing", false)
	repo.PropertySets = d.getSet("property_sets")
	repo.HandleReleases = d.getBoolRef("handle_releases", false)
	repo.HandleSnapshots = d.getBoolRef("handle_snapshots", false)
	repo.ChecksumPolicyType = d.getString("checksum_policy_type", false)
	repo.MaxUniqueSnapshots = d.getInt("max_unique_snapshots", false)
	repo.SnapshotVersionBehavior = d.getString("snapshot_version_behavior", false)
	repo.SuppressPomConsistencyChecks = d.getBoolRef("suppress_pom_consistency_checks", false)
	repo.XrayIndex = d.getBoolRef("xray_index", false)
	repo.ForceNugetAuthentication = d.getBool("force_nuget_authentication", false)

	return repo
}

func resourceLocalRepositoryCreate(d *schema.ResourceData, m interface{}) error {

	client := m.(*ArtClient).Resty

	repo := unmarshalLocalRepository(d)

	_, err := client.R().SetBody(repo).
		SetHeader("accept", "text/plain").
		SetBody(repo).Put("/artifactory/api/repositories/" + repo.Key)

	if err != nil {
		return err
	}
	d.SetId(repo.Key)
	return resourceLocalRepositoryRead(d, m)
}

func resourceLocalRepositoryRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*ArtClient).Resty

	repo := MessyRepo{}
	if d.Id() == "" {
		return fmt.Errorf("no id given")
	}

	resp, err := client.R().SetResult(&repo).Get("/artifactory/api/repositories/" + d.Id())
	if err != nil {
		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error: id: %s %s %s", d.Id(), err, string(resp.Body()[:]))
	}

	hasErr := false
	logError := cascadingErr(&hasErr)

	logError(d.Set("key", repo.Key))
	// type 'yum' is not to be supported, as this is really of type 'rpm'. When 'yum' is used on create, RT will
	// respond with 'rpm' and thus confuse TF into think there has been a state change.
	logError(d.Set("package_type", repo.PackageType))
	logError(d.Set("description", repo.Description))
	logError(d.Set("notes", repo.Notes))
	logError(d.Set("includes_pattern", repo.IncludesPattern))
	logError(d.Set("excludes_pattern", repo.ExcludesPattern))
	logError(d.Set("repo_layout_ref", repo.RepoLayoutRef))
	logError(d.Set("debian_trivial_layout", repo.DebianTrivialLayout))
	logError(d.Set("max_unique_tags", repo.MaxUniqueTags))
	logError(d.Set("blacked_out", repo.BlackedOut))
	logError(d.Set("archive_browsing_enabled", repo.ArchiveBrowsingEnabled))
	logError(d.Set("calculate_yum_metadata", repo.CalculateYumMetadata))
	logError(d.Set("yum_root_depth", repo.YumRootDepth))
	logError(d.Set("docker_api_version", repo.DockerApiVersion))
	logError(d.Set("enable_file_lists_indexing", repo.EnableFileListsIndexing))
	logError(d.Set("property_sets", schema.NewSet(schema.HashString, castToInterfaceArr(repo.PropertySets))))
	logError(d.Set("handle_releases", repo.HandleReleases))
	logError(d.Set("handle_snapshots", repo.HandleSnapshots))
	logError(d.Set("checksum_policy_type", repo.ChecksumPolicyType))
	logError(d.Set("max_unique_snapshots", repo.MaxUniqueSnapshots))
	logError(d.Set("snapshot_version_behavior", repo.SnapshotVersionBehavior))
	logError(d.Set("suppress_pom_consistency_checks", repo.SuppressPomConsistencyChecks))
	logError(d.Set("xray_index", repo.XrayIndex))
	logError(d.Set("force_nuget_authentication", repo.ForceNugetAuthentication))

	if hasErr {
		return fmt.Errorf("failed saving state for local repos")
	}

	return nil
}

func resourceLocalRepositoryUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ArtClient).Resty

	repo := unmarshalLocalRepository(d)

	_, err := client.R().SetBody(repo).SetHeader("accept", "text/plain").
		Post("/artifactory/api/repositories/" + d.Id())

	if err != nil {
		return err
	}
	d.SetId(repo.Key) // I have no idea why someone would do this when d.Id() is already correct
	return resourceLocalRepositoryRead(d, m)
}

func resourceLocalRepositoryDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*ArtClient).Resty

	_, err := client.R().SetHeader("accept", "*/*").Delete("/artifactory/api/repositories/" + d.Id())
	return err
}

func resourceLocalRepositoryExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*ArtClient).Resty

	_, err := client.R().Head("/artifactory/api/repositories/" + d.Id())
	// artifactory returns 400 instead of 404. but regardless, it's an error
	return err == nil, err
}
