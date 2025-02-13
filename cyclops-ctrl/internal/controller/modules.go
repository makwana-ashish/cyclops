package controller

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/cyclops-ui/cycops-ctrl/api/v1alpha1"
	"github.com/cyclops-ui/cycops-ctrl/internal/cluster/k8sclient"
	"github.com/cyclops-ui/cycops-ctrl/internal/mapper"
	"github.com/cyclops-ui/cycops-ctrl/internal/models/dto"
	"github.com/cyclops-ui/cycops-ctrl/internal/storage/templates"
	"github.com/cyclops-ui/cycops-ctrl/internal/template"
	"github.com/gin-gonic/gin"
)

type Modules struct {
	kubernetesClient *k8sclient.KubernetesClient
	templates        *templates.Storage
}

func NewModulesController(templates *templates.Storage, kubernetes *k8sclient.KubernetesClient) *Modules {
	return &Modules{
		kubernetesClient: kubernetes,
		templates:        templates,
	}
}

func (m *Modules) GetModule(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	module, err := m.kubernetesClient.GetModule(ctx.Param("name"))
	if err != nil {
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	moduleDTO, err := mapper.ModuleToDTO(*module)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error mapping module", err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, moduleDTO)
}

func (m *Modules) ListModules(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	modules, err := m.kubernetesClient.ListModules()
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error fetching modules", err.Error()))
		return
	}

	dtoModules := mapper.ModuleListToDTO(modules)

	for i, dtoModule := range dtoModules {
		dtoModuleStatus, err := m.kubernetesClient.GetModuleResourcesHealth(dtoModule.Name)
		if err != nil {
			fmt.Println(err)
			ctx.JSON(http.StatusInternalServerError, dto.NewError("Error fetching modules", err.Error()))
			return
		}

		dtoModules[i].Status = dtoModuleStatus
	}

	ctx.JSON(http.StatusOK, dtoModules)
}

func (m *Modules) DeleteModule(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	err := m.kubernetesClient.DeleteModule(ctx.Param("name"))
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error deleting module", err.Error()))
		return
	}

	ctx.Status(http.StatusOK)
}

func (m *Modules) GetModuleHistory(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	module, err := m.kubernetesClient.GetModule(ctx.Param("name"))
	if err != nil {
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, module.History)
}

func (m *Modules) Manifest(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	var request v1alpha1.ModuleSpec
	if err := ctx.BindJSON(&request); err != nil {
		fmt.Println("error binding request", request)
		ctx.JSON(http.StatusBadRequest, dto.NewError("Error loading template", err.Error()))
		return
	}

	targetTemplate, err := template.GetTemplate(
		request.TemplateRef.URL,
		request.TemplateRef.Path,
		request.TemplateRef.Version,
	)
	if err != nil {
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	manifest, err := template.HelmTemplate(v1alpha1.Module{Spec: request}, targetTemplate)
	if err != nil {
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	manifest = strings.TrimPrefix(manifest, "\n---")
	manifest = strings.TrimSuffix(manifest, "---\n")

	ctx.String(http.StatusOK, manifest)
}

func (m *Modules) CurrentManifest(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	module, err := m.kubernetesClient.GetModule(ctx.Param("name"))
	if err != nil {
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	targetTemplate, err := template.GetTemplate(
		module.Spec.TemplateRef.URL,
		module.Spec.TemplateRef.Path,
		module.Spec.TemplateRef.Version,
	)
	if err != nil {
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	manifest, err := template.HelmTemplate(*module, targetTemplate)
	if err != nil {
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	manifest = strings.TrimPrefix(manifest, "---\n")
	manifest = strings.TrimSuffix(manifest, "---\n")

	ctx.String(http.StatusOK, manifest)
}

func (m *Modules) DeleteModuleResource(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	var request dto.DeleteResource
	if err := ctx.BindJSON(&request); err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusBadRequest, dto.NewError("Error mapping module request", err.Error()))
		return
	}

	err := m.kubernetesClient.Delete(&request)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error deleting module", err.Error()))
		return
	}

	ctx.Status(http.StatusOK)
}

func (m *Modules) CreateModule(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	var request dto.Module
	if err := ctx.BindJSON(&request); err != nil {
		fmt.Println("error binding request", request)
		ctx.JSON(http.StatusBadRequest, dto.NewError("Error loading template", err.Error()))
		return
	}

	module, err := mapper.RequestToModule(request)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error mapping module", err.Error()))
		return
	}

	err = m.kubernetesClient.CreateModule(module)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error creating module", err.Error()))
		return
	}

	ctx.Status(http.StatusOK)
}

func (m *Modules) UpdateModule(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	var request dto.Module
	if err := ctx.BindJSON(&request); err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusBadRequest, dto.NewError("Error mapping module request", err.Error()))
		return
	}

	curr, err := m.kubernetesClient.GetModule(request.Name)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error fetcing module", err.Error()))
		return
	}

	module, err := mapper.RequestToModule(request)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error creating module", err.Error()))
		return
	}

	history := curr.History
	if curr.History == nil {
		history = make([]v1alpha1.HistoryEntry, 0)
	}

	module.History = append([]v1alpha1.HistoryEntry{{
		Generation:  curr.Generation,
		TemplateRef: curr.Spec.TemplateRef,
		Values:      curr.Spec.Values,
	}}, history...)

	if len(module.History) > 10 {
		module.History = module.History[:len(module.History)-1]
	}

	module.SetResourceVersion(curr.GetResourceVersion())

	err = m.kubernetesClient.UpdateModule(module)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error updating module", err.Error()))
		return
	}

	ctx.Status(http.StatusOK)
}

func (m *Modules) ResourcesForModule(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	module, err := m.kubernetesClient.GetModule(ctx.Param("name"))
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusBadRequest, dto.NewError("Error mapping module request", err.Error()))
		return
	}

	t, err := template.GetTemplate(
		module.Spec.TemplateRef.URL,
		module.Spec.TemplateRef.Path,
		module.Spec.TemplateRef.Version,
	)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error fetching template", err.Error()))
		return
	}

	resources, err := m.kubernetesClient.GetResourcesForModule(ctx.Param("name"))
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error fetching module resources", err.Error()))
		return
	}

	resources, err = m.kubernetesClient.GetDeletedResources(resources, *module, t)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error fetching deleted module resources", err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, resources)
}

func (m *Modules) Template(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	module, err := m.kubernetesClient.GetModule(ctx.Param("name"))
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error fetching module", err.Error()))
		return
	}

	currentTemplate, err := template.GetTemplate(
		module.Spec.TemplateRef.URL,
		module.Spec.TemplateRef.Path,
		module.Spec.TemplateRef.Version,
	)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error fetching template", err.Error()))
		return
	}

	currentManifest, err := template.HelmTemplate(*module, currentTemplate)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error templating current", err.Error()))
		return
	}

	proposedTemplate, err := template.GetTemplate(
		module.Spec.TemplateRef.URL,
		module.Spec.TemplateRef.Path,
		module.Spec.TemplateRef.Version,
	)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error creating proposed template", err.Error()))
		return
	}

	proposedManifest, err := template.HelmTemplate(*module, proposedTemplate)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error templating proposed", err.Error()))
		return
	}

	res := dto.TemplatesResponse{
		Current: currentManifest,
		New:     proposedManifest,
	}

	ctx.JSON(http.StatusOK, res)
}

func (m *Modules) HelmTemplate(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	module, err := m.kubernetesClient.GetModule(ctx.Param("name"))
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error fetching module", err.Error()))
		return
	}

	currentTemplate, err := template.GetTemplate(
		module.Spec.TemplateRef.URL,
		module.Spec.TemplateRef.Path,
		module.Spec.TemplateRef.Version,
	)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error fetching template", err.Error()))
		return
	}

	_, err = template.HelmTemplate(*module, currentTemplate)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error templating", err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, "{}")
}

//func (m *Modules) ModuleToResources(ctx *gin.Context) {
//	err := m.kubernetesClient.ModuleToResources("test")
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	ctx.Header("Access-Control-Allow-Origin", "*")
//	ctx.Status(http.StatusOK)
//}
//
//func (m *Modules) ResourcesForModule(ctx *gin.Context) {
//	resources, err := m.kubernetesClient.ResourcesForModule(ctx.Param("name"))
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	ctx.Header("Access-Control-Allow-Origin", "*")
//	ctx.JSON(http.StatusOK, resources)
//}

func (m *Modules) GetLogs(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	logCount := int64(100)
	logs, err := m.kubernetesClient.GetPodLogs(
		ctx.Param("namespace"),
		ctx.Param("container"),
		ctx.Param("name"),
		&logCount,
	)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error fetching logs", err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, logs)
}

func (m *Modules) DownloadLogs(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	namespace := ctx.Param("namespace")
	container := ctx.Param("container")
	name := ctx.Param("name")

	logs, err := m.kubernetesClient.GetPodLogs(
		namespace,
		container,
		name,
		nil,
	)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error fetching logs", err.Error()))
		return
	}

	tempFile, err := os.CreateTemp("", fmt.Sprintf("%v-%v-*.txt", name, container))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
		return
	}
	defer tempFile.Close()

	for _, log := range logs {
		_, err = tempFile.WriteString(log)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write to file"})
			return
		}
	}

	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Disposition", "attachment; filename="+fmt.Sprintf("%v-%v.txt", name, container))
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.File(tempFile.Name())
}

func (m *Modules) GetManifest(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	group := ctx.Query("group")
	version := ctx.Query("version")
	kind := ctx.Query("kind")
	name := ctx.Query("name")
	namespace := ctx.Query("namespace")

	manifest, err := m.kubernetesClient.GetManifest(group, version, kind, name, namespace)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to fetch resource manifest",
			"reason": err.Error(),
		})
		return
	}

	ctx.String(http.StatusOK, manifest)
}

func (m *Modules) GetResource(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	group := ctx.Query("group")
	version := ctx.Query("version")
	kind := ctx.Query("kind")
	name := ctx.Query("name")
	namespace := ctx.Query("namespace")

	resource, err := m.kubernetesClient.GetResource(group, version, kind, name, namespace)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to fetch resource",
			"reason": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, resource)
}

func getTargetGeneration(generation string, module *v1alpha1.Module) (*v1alpha1.Module, bool) {
	// no generation specified means current generation
	if len(generation) == 0 {
		return module, true
	}

	var target *v1alpha1.HistoryEntry
	for _, entry := range module.History {
		if fmt.Sprintf("%v", entry.Generation) == generation {
			target = &entry
		}
	}

	if target == nil {
		return nil, false
	}

	return &v1alpha1.Module{
		TypeMeta:   module.TypeMeta,
		ObjectMeta: module.ObjectMeta,
		Spec:       module.Spec,
		Status:     module.Status,
	}, true
}
