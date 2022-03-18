package main

import (
	"fmt"
	"github.com/golang/glog"
	tfv1 "github.com/isaaguilar/terraform-operator/pkg/apis/tf/v1alpha1"
	tfclientset "github.com/isaaguilar/terraform-operator/pkg/client/clientset/versioned"
	"github.com/labstack/gommon/color"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	//"github.com/golang/glog"
	"github.com/labstack/echo"
	"os"

	"html/template"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345
	"k8s.io/client-go/tools/clientcmd"

	//"os"
	//"strconv"
)

var kubeconfig string

// TemplateRenderer is a custom html/template renderer for Echo framework
type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {

	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	kubeconfig = os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		glog.Fatalf("Cannot get kubeconfig from: %v", "KUBECONFIG")
	}

	e := echo.New()

	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
	e.Renderer = renderer

	e.File("/img/tf.png", "img/tf.png")
	e.File("/img/running.png", "img/running.png")
	e.File("/img/pending.png", "img/pending.png")
	e.File("/img/completed.png", "img/completed.png")
	e.File("/img/deleting.png", "img/deleting.png")
	e.File("/img/fail.png", "img/fail.png")
	e.File("/", "index.html")
	e.File("/create", "views/create.html")
	e.GET("/get", Get)
	e.GET("/list", List)
	//e.GET("/templates", Templates)
	//e.GET("/spaces", Spaces)
	//e.GET("/services", Services)
	e.GET("/createnew", CreateNew)
	e.GET("/edit", Edit)
	e.GET("/getedit", GetEdit)
	e.GET("/delete", Delete)
	e.GET("/logs", Logs)
	e.Logger.Fatal(e.Start(":1323"))
}


func Get(c echo.Context) error {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %v", err)
	}

	tfClient, err := tfclientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building tf clientset: %v", err)
	}

	tf, err := tfClient.TfV1alpha1().Terraforms("default").Get(context.TODO(), c.Request().FormValue("name"), metav1.GetOptions{})
	if err != nil {
		glog.Fatalf("Error getting terraform resource: %v", c.Request().FormValue("name"))
	}
	fmt.Print("Get terraform resource", tf.Name)
	return c.Render(http.StatusOK, "get.html", tf)
}

func Delete(c echo.Context) error {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %v", err)
	}

	tfClient, err := tfclientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building tf clientset: %v", err)
	}

	tf, err := tfClient.TfV1alpha1().Terraforms("default").Get(context.TODO(), c.Request().FormValue("name"), metav1.GetOptions{})
	if err != nil {
		glog.Fatalf("Error getting terraform resource: %v", c.Request().FormValue("name"))
	}
	fmt.Print("Get terraform resource", tf.Name)

	err = tfClient.TfV1alpha1().Terraforms("default").Delete(context.TODO(), tf.Name, metav1.DeleteOptions{})
	if err != nil {
		glog.Fatalf("Error deleting terraform resource: %v", c.Request().FormValue("name"))
	}
	fmt.Print("Delete terraform resource", tf.Name)
	return c.Render(http.StatusOK, "deleteDone.html", tf.Name)
}

func List(c echo.Context) error {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %v", err)
	}

	tfClient, err := tfclientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building knap clientset: %v", err)
	}

	tfLst, err := tfClient.TfV1alpha1().Terraforms("default").List(context.TODO(), metav1.ListOptions{})
	color.Cyan("%-30s%-20s%-20s%-20s%-20s\n", "Terraform Name", "Namespace", "Version", "Creation Time", "Phase")
	for i := 0; i < len(tfLst.Items); i++ {
		tf := tfLst.Items[i]
		fmt.Printf("%-30s%-20s%-20s%-20s%-20s\n", tf.Name, tf.Namespace, tf.Generation, tf.CreationTimestamp, tf.Status.Phase)
	}
	return c.Render(http.StatusOK, "list.html", tfLst.Items)
}

//func Templates(c echo.Context) error {
//	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
//	if err != nil {
//		glog.Fatalf("Error building kubeconfig: %v", err)
//	}
//
//	tektoncdClient, err := tektoncdclientset.NewForConfig(cfg)
//	if err != nil {
//		glog.Fatalf("Error building tektoncd clientset: %v", err)
//	}
//
//	pipelines, err := tektoncdClient.TektonV1alpha1().Pipelines("default").List(metav1.ListOptions{})
//	color.Cyan("%-40s%-80s\n", "Template Name", "Template Flow")
//	for i := 0; i < len(pipelines.Items); i++ {
//		pipeline := pipelines.Items[i]
//		taskFlow := ""
//		for i := 0; i < len(pipeline.Spec.Tasks); i++ {
//			task := pipeline.Spec.Tasks[i]
//			if taskFlow == "" {
//				taskFlow = task.Name
//			} else {
//				taskFlow = taskFlow + " -> " + task.Name
//			}
//		}
//		fmt.Printf("%-40s%-80s\n", pipeline.Name, taskFlow)
//	}
//	return c.Render(http.StatusOK, "templates.html", pipelines.Items)
//}
//
//func Spaces(c echo.Context) error {
//	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
//	if err != nil {
//		glog.Fatalf("Error building kubeconfig: %v", err)
//	}
//
//	knapClient, err := knapclientset.NewForConfig(cfg)
//	if err != nil {
//		glog.Fatalf("Error building knap clientset: %v", err)
//	}
//
//	appLst, err := knapClient.KnapV1alpha1().Appengines("default").List(metav1.ListOptions{})
//	color.Cyan("%-30s%-20s%-20s%-20s%-20s\n", "Engine Name", "Application Name", "Ready", "Instance", "Domain")
//	for i := 0; i < len(appLst.Items); i++ {
//		app := appLst.Items[i]
//		fmt.Printf("%-30s%-20s%-20s%-20s%-20s\n", app.Name, app.Spec.AppName, app.Status.Ready, fmt.Sprint(app.Status.Instance)+"/"+fmt.Sprint(app.Spec.Size), app.Status.Domain)
//	}
//	return c.Render(http.StatusOK, "spaces.html", appLst.Items)
//}
//
//func Services(c echo.Context) error {
//	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
//	if err != nil {
//		glog.Fatalf("Error building kubeconfig: %v", err)
//	}
//
//	knapClient, err := knapclientset.NewForConfig(cfg)
//	if err != nil {
//		glog.Fatalf("Error building knap clientset: %v", err)
//	}
//
//	appLst, err := knapClient.KnapV1alpha1().Appengines("default").List(metav1.ListOptions{})
//	color.Cyan("%-30s%-20s%-20s%-20s%-20s\n", "Engine Name", "Application Name", "Ready", "Instance", "Domain")
//	for i := 0; i < len(appLst.Items); i++ {
//		app := appLst.Items[i]
//		fmt.Printf("%-30s%-20s%-20s%-20s%-20s\n", app.Name, app.Spec.AppName, app.Status.Ready, fmt.Sprint(app.Status.Instance)+"/"+fmt.Sprint(app.Spec.Size), app.Status.Domain)
//	}
//	return c.Render(http.StatusOK, "services.html", appLst.Items)
//}

func CreateNew(c echo.Context) error {
	r := c.Request()
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %v", err)
	}

	tfClient, err := tfclientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building tf clientset: %v", err)
	}

	tf := &tfv1.Terraform{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.FormValue("tfName"),
			Namespace: r.FormValue("namespace"),
		},
		Spec: tfv1.TerraformSpec{
			CustomBackend: fmt.Sprintf(`terraform {
      backend "kubernetes" {
        secret_suffix    = "%s"
        in_cluster_config  = true
      }
    }
`, r.FormValue("tfName")),
			TerraformVersion: r.FormValue("tfVersion"),
			TerraformModule:     r.FormValue("gitRepo"),
			TerraformRunnerPullPolicy: corev1.PullIfNotPresent,
			KeepCompletedPods: true,
			WriteOutputsToStatus: true,
			Env: []corev1.EnvVar{
				{
					Name:      r.FormValue("envName"),
					Value:     r.FormValue("envValue"),
				},
			},
		},
	}
	_, err = tfClient.TfV1alpha1().Terraforms("default").Create(context.TODO(), tf, metav1.CreateOptions{})

	if err != nil {
		//glog.Fatalf("Error creating application engine: %s", args[0])
		fmt.Println("Error creating terraform resource", r.FormValue("tfName"), err)
	} else {
		fmt.Println("Terraform Resource", r.FormValue("tfName"), "is created successfully")
	}

	return c.Render(http.StatusOK, "createDone.html", map[string]interface{}{
		"name": r.FormValue("tfName"),
	})
}

func Edit(c echo.Context) error {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %v", err)
	}

	tfClient, err := tfclientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building tf clientset: %v", err)
	}

	tf, err := tfClient.TfV1alpha1().Terraforms("default").Get(context.TODO(), c.Request().FormValue("name"), metav1.GetOptions{})
	if err != nil {
		glog.Fatalf("Error getting terraform resource: %v", c.Request().FormValue("name"))
	}
	fmt.Print("Get terraform resource", tf.Name)
	return c.Render(http.StatusOK, "edit.html", tf)
}

func GetEdit(c echo.Context) error {
	r := c.Request()
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %v", err)
	}

	tfClient, err := tfclientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building tf clientset: %v", err)
	}

	tf, err := tfClient.TfV1alpha1().Terraforms("default").Get(context.TODO(), r.FormValue("appName")+"-appengine", metav1.GetOptions{})
	if err != nil {
		glog.Fatalf("Error getting terraform resource: %v", r.FormValue("tfName"))
	}

	tf.Spec.TerraformModule = r.FormValue("gitRepo")

	tf.Spec.Env[0].Name = r.FormValue("envName")
	tf.Spec.Env[0].Value = r.FormValue("envValue")


	_, err = tfClient.TfV1alpha1().Terraforms("default").Update(context.TODO(), tf, metav1.UpdateOptions{})

	if err != nil {
		//glog.Fatalf("Error creating terraform resource: %s", args[0])
		fmt.Println("Error updating terraform resource", r.FormValue("tfName"), err)
	} else {
		fmt.Println("Terraform Resource", r.FormValue("tfName"), "is updated successfully")
	}

	return c.Render(http.StatusOK, "editDone.html", map[string]interface{}{
		"name": r.FormValue("tfName"),
	})
}

func Logs(c echo.Context) error {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %v", err)
	}

	tfClient, err := tfclientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building tf clientset: %v", err)
	}

	tf, err := tfClient.TfV1alpha1().Terraforms("default").Get(context.TODO(), c.Request().FormValue("name"), metav1.GetOptions{})
	if err != nil {
		glog.Fatalf("Error getting terraform resource: %v", c.Request().FormValue("name"))
	}
	fmt.Print("Get terraform resource", tf)
	return c.Render(http.StatusOK, "logs.html", tf)
}