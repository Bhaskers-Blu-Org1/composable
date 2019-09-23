/*
 * Copyright 2019 IBM Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controllers

import (
	//	"github.com/ibm/composable/api"
	"github.com/ibm/composable/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	//	"k8s.io/client-go/kubernetes/scheme"
	//	"k8s.io/client-go/rest"
	"k8s.io/klog"
	//	"sigs.k8s.io/controller-runtime/pkg/envtest"
	//	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	testContext test.TestContext
	//testEnv     *envtest.Environment
	stop chan struct{}
)

//func TestComposable(t *testing.T) {
//	klog.InitFlags(flag.CommandLine)
//	klog.SetOutput(GinkgoWriter)
//
//	RegisterFailHandler(Fail)
//	SetDefaultEventuallyPollingInterval(1 * time.Second)
//	SetDefaultEventuallyTimeout(60 * time.Second)
//
//	RunSpecs(t, "Composable Suite")
//}

/*
var _ = BeforeSuite(func() {

	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "config", "crds"),
			filepath.Join("./testdata", "crds")},
		ControlPlaneStartTimeout: 2 * time.Minute,
	}
	api.AddToScheme(scheme.Scheme)

	var err error
	var cfg *rest.Config
	if cfg, err = testEnv.Start(); err != nil {
		log.Fatal(err)
	}

	syncPeriod := 30 * time.Second // set a sync period
	mgr, err := manager.New(cfg, manager.Options{SyncPeriod: &syncPeriod})
	Expect(err).NotTo(HaveOccurred())

	client := mgr.GetClient()

	recFn := newReconciler(mgr)
	Expect(add(mgr, recFn)).NotTo(HaveOccurred())
	stop = test.StartTestManager(mgr)
	testNs := test.SetupKubeOrDie(cfg, "test-ns-")
	testContext = test.NewTestContext(client, testNs)

})

var _ = AfterSuite(func() {
	close(stop)
	testEnv.Stop()
})
*/

var _ = Describe("test Composable operations", func() {
	dataDir := "testdata/"
	unstrObj := unstructured.Unstructured{}

	strArray := []interface{}{"kafka01-prod02.messagehub.services.us-south.bluemix.net:9093",
		"kafka02-prod02.messagehub.services.us-south.bluemix.net:9093",
		"kafka03-prod02.messagehub.services.us-south.bluemix.net:9093",
		"kafka04-prod02.messagehub.services.us-south.bluemix.net:9093",
		"kafka05-prod02.messagehub.services.us-south.bluemix.net:9093"}

	AfterEach(func() {
		// delete the Composable object
		comp := test.LoadCompasable(dataDir + "compCopy.yaml")
		test.DeleteInNs(testContext, &comp, false)
		Eventually(test.GetObject(testContext, &comp)).Should(BeNil())

		obj := test.LoadObject(dataDir+"inputDataObject.yaml", &unstructured.Unstructured{})
		test.DeleteObject(testContext, obj, false)
		Eventually(test.GetObject(testContext, obj)).Should(BeNil())
	})

	It("Composable should successfully set default values to the output object", func() {
		By("Deploy Composable object")
		comp := test.LoadCompasable(dataDir + "compCopy.yaml")
		test.PostInNs(testContext, &comp, false, 0)
		Eventually(test.GetObject(testContext, &comp)).ShouldNot(BeNil())

		By("Get Output object")
		groupVersionKind := schema.GroupVersionKind{Kind: "OutputValue", Version: "v1", Group: "test.ibmcloud.ibm.com"}
		unstrObj.SetGroupVersionKind(groupVersionKind)
		objNamespacedname := types.NamespacedName{Namespace: testContext.Namespace(), Name: "comp-out"}
		klog.V(5).Infof("Get Object %s\n", objNamespacedname)
		Eventually(test.GetUnstructuredObject(testContext, objNamespacedname, &unstrObj)).Should(Succeed())
		testSpec, ok := unstrObj.Object[spec].(map[string]interface{})
		Ω(ok).Should(BeTrue())

		By("default intValue")
		Ω(testSpec["intValue"]).Should(BeEquivalentTo(10))

		By("default floatValue")
		Ω(testSpec["floatValue"]).Should(BeEquivalentTo(10.1))

		By("default boolValue")
		Ω(testSpec["boolValue"]).Should(BeFalse())

		By("default stringValue")
		Ω(testSpec["stringValue"]).Should(Equal("default"))

		By("default stringFromBase64")
		Ω(testSpec["stringFromBase64"]).Should(Equal("default"))

		By("default arrayStrings")
		Ω(testSpec["arrayStrings"]).Should(BeEquivalentTo([]interface{}{"aa", "bb", "cc"}))

		By("default arrayIntegers")
		Ω(testSpec["arrayIntegers"]).Should(BeEquivalentTo([]interface{}{int64(1), int64(0), int64(1)}))

		By("default objectValue")
		Ω(testSpec["objectValue"]).Should(BeEquivalentTo(map[string]interface{}{"family": "DefaultFamilyName", "first": "DefaultFirstName", "age": int64(-1)}))

		By("default stringJson2Value")
		Ω(testSpec["stringJson2Value"]).Should(BeEquivalentTo("default1,default2,default3"))

	})

	It("Composable should successfully copy values to the output object", func() {

		By("Deploy input Object")
		obj := test.LoadObject(dataDir+"inputDataObject.yaml", &unstructured.Unstructured{})
		test.CreateObject(testContext, obj, false, 0)
		Eventually(test.GetObject(testContext, obj)).ShouldNot(BeNil())

		groupVersionKind := schema.GroupVersionKind{Kind: "InputValue", Version: "v1", Group: "test.ibmcloud.ibm.com"}
		unstrObj.SetGroupVersionKind(groupVersionKind)
		objNamespacedname := types.NamespacedName{Namespace: "default", Name: "inputdata"}
		klog.V(5).Infof("Get Object %s\n", objNamespacedname)
		Eventually(test.GetUnstructuredObject(testContext, objNamespacedname, &unstrObj)).Should(Succeed())

		By("Deploy Composable object")
		comp := test.LoadCompasable(dataDir + "compCopy.yaml")
		test.PostInNs(testContext, &comp, false, 0)
		Eventually(test.GetObject(testContext, &comp)).ShouldNot(BeNil())
		Eventually(test.GetState(testContext, &comp)).Should(Equal(OnlineStatus))

		By("Get Output object")
		groupVersionKind = schema.GroupVersionKind{Kind: "OutputValue", Version: "v1", Group: "test.ibmcloud.ibm.com"}
		unstrObj.SetGroupVersionKind(groupVersionKind)
		objNamespacedname = types.NamespacedName{Namespace: testContext.Namespace(), Name: "comp-out"}
		klog.V(5).Infof("Get Object %s\n", objNamespacedname)
		Eventually(test.GetUnstructuredObject(testContext, objNamespacedname, &unstrObj)).Should(Succeed())
		testSpec, ok := unstrObj.Object[spec].(map[string]interface{})
		Ω(ok).Should(BeTrue())

		By("copy intValue")
		// We use Eventually so the controller will be able to update teh destination object
		Ω(testSpec["intValue"]).Should(BeEquivalentTo(12))
		//
		By("copy floatValue")
		Ω(testSpec["floatValue"].(float64)).Should(BeEquivalentTo(23.5))

		By("copy boolValue")
		Ω(testSpec["boolValue"]).Should(BeTrue())

		By("copy stringValue")
		Ω(testSpec["stringValue"]).Should(Equal("Hello world"))

		By("copy stringFromBase64")
		Ω(testSpec["stringFromBase64"]).Should(Equal("9376"))

		By("copy arrayStrings")
		Ω(testSpec["arrayStrings"]).Should(Equal(strArray))

		By("copy arrayIntegers")
		Ω(testSpec["arrayIntegers"]).Should(Equal([]interface{}{int64(1), int64(2), int64(3), int64(4)}))

		By("copy objectValue")
		Ω(testSpec["objectValue"]).Should(Equal(map[string]interface{}{"family": "FamilyName", "first": "FirstName", "age": int64(27)}))

		By("copy stringJson2Value")
		val, _ := Array2CSStringTransformer(strArray)
		Ω(testSpec["stringJson2Value"]).Should(BeEquivalentTo(val))

	})
	It("Composable should successfully update values of the output object", func() {

		gvkIn := schema.GroupVersionKind{Kind: "InputValue", Version: "v1", Group: "test.ibmcloud.ibm.com"}
		gvkOut := schema.GroupVersionKind{Kind: "OutputValue", Version: "v1", Group: "test.ibmcloud.ibm.com"}
		objNamespacednameIn := types.NamespacedName{Namespace: "default", Name: "inputdata"}
		objNamespacednameOut := types.NamespacedName{Namespace: testContext.Namespace(), Name: "comp-out"}

		//unstrObj.SetGroupVersionKind(gvkOut)
		// First, the output object is created with default values, after that we deploy the inputObject and will check
		// that all Output object filed are updated.
		By("check that input object doesn't exist") // the object should not exist
		unstrObj.SetGroupVersionKind(gvkIn)
		Ω(test.GetUnstructuredObject(testContext, objNamespacednameIn, &unstrObj)()).Should(HaveOccurred())

		By("check that output object doesn't exist. If it does => remove it ") // the object should not exist, or we delete it
		unstrObj.SetGroupVersionKind(gvkOut)
		err2 := test.GetUnstructuredObject(testContext, objNamespacednameOut, &unstrObj)()
		if err2 == nil {
			test.DeleteObject(testContext, &unstrObj, false)
			Eventually(test.GetObject(testContext, &unstrObj)).Should(BeNil())
		}

		By("deploy Composable object")
		comp := test.LoadCompasable(dataDir + "compCopy.yaml")
		test.PostInNs(testContext, &comp, false, 0)
		Eventually(test.GetObject(testContext, &comp)).ShouldNot(BeNil())

		By("get Output object")
		unstrObj.SetGroupVersionKind(gvkOut)
		Eventually(test.GetUnstructuredObject(testContext, objNamespacednameOut, &unstrObj)).Should(Succeed())
		testSpec, ok := unstrObj.Object[spec].(map[string]interface{})
		Ω(ok).Should(BeTrue())

		// Check some of default the values
		By("check default intValue")
		Ω(testSpec["intValue"]).Should(BeEquivalentTo(10))

		By("check default stringFromBase64")
		Ω(testSpec["stringFromBase64"]).Should(Equal("default"))

		By("deploy input Object")
		obj := test.LoadObject(dataDir+"inputDataObject.yaml", &unstructured.Unstructured{})
		test.CreateObject(testContext, obj, false, 0)
		Eventually(test.GetObject(testContext, obj)).ShouldNot(BeNil())

		By("check updated inValue")
		unstrObj = unstructured.Unstructured{}
		unstrObj.SetGroupVersionKind(gvkOut)
		Eventually(func() (int64, error) {
			err := test.GetUnstructuredObject(testContext, objNamespacednameOut, &unstrObj)()
			if err != nil {
				return int64(0), err
			}
			testSpec, _ = unstrObj.Object[spec].(map[string]interface{})
			return testSpec["intValue"].(int64), nil
		}).Should(Equal(int64(12)))

		// Check other values
		By("check updated floatValue")
		Ω(testSpec["floatValue"].(float64)).Should(BeEquivalentTo(23.5))

		By("check updated boolValue")
		Ω(testSpec["boolValue"]).Should(BeTrue())

		By("check updated stringValue")
		Ω(testSpec["stringValue"]).Should(Equal("Hello world"))

		By("check updated stringFromBase64")
		Ω(testSpec["stringFromBase64"]).Should(Equal("9376"))

		By("check updated arrayStrings")
		Ω(testSpec["arrayStrings"]).Should(Equal(strArray))

		By("check updated arrayIntegers")
		Ω(testSpec["arrayIntegers"]).Should(Equal([]interface{}{int64(1), int64(2), int64(3), int64(4)}))

		By("check updated objectValue")
		Ω(testSpec["objectValue"]).Should(Equal(map[string]interface{}{"family": "FamilyName", "first": "FirstName", "age": int64(27)}))

		By("check updated stringJson2Value")
		val, _ := Array2CSStringTransformer(strArray)
		Ω(testSpec["stringJson2Value"]).Should(BeEquivalentTo(val))
	})
})

var _ = Describe("Validate group separation", func() {
	Context("There are 3 groups that have Kind = `Service`. They are: Service/v1; Service.ibmcloud.ibm.com/v1alpha1 and Service.test.ibmcloud.ibm.com/v1", func() {
		It("Composable should correctly discover required objects", func() {
			dataDir := "testdata/"

			By("deploy K8s Service")
			kubeObj := test.LoadObject(dataDir+"serviceK8s.yaml", &v1.Service{})
			test.CreateObject(testContext, kubeObj, false, 0)
			Eventually(test.GetObject(testContext, kubeObj)).ShouldNot(BeNil())

			By("deploy test Service")
			tObj := test.LoadObject(dataDir+"serviceTest.yaml", &unstructured.Unstructured{})
			test.CreateObject(testContext, tObj, false, 0)
			Eventually(test.GetObject(testContext, tObj)).ShouldNot(BeNil())

			By("deploy Composable object")
			comp := test.LoadCompasable(dataDir + "compServices.yaml")
			test.PostInNs(testContext, &comp, false, 0)
			Eventually(test.GetObject(testContext, &comp)).ShouldNot(BeNil())

			By("get the output object and validate its fields")
			unstrObj := unstructured.Unstructured{}
			gvk := schema.GroupVersionKind{Kind: "OutputValue", Version: "v1", Group: "test.ibmcloud.ibm.com"}
			objNamespacedname := types.NamespacedName{Namespace: testContext.Namespace(), Name: "services-out"}

			unstrObj.SetGroupVersionKind(gvk)
			Eventually(test.GetUnstructuredObject(testContext, objNamespacedname, &unstrObj)).Should(Succeed())
			testSpec, ok := unstrObj.Object[spec].(map[string]interface{})
			Ω(ok).Should(BeTrue())

			Ω(testSpec["k8sValue"]).Should(Equal("None"))
			Ω(testSpec["testValue"]).Should(Equal("Test"))

		})

	})
})

var _ = Describe("IBM cloud-operators compatibility", func() {
	dataDir := "testdata/cloud-operators-data/"
	groupVersionKind := schema.GroupVersionKind{Kind: "Service", Version: "v1alpha1", Group: "ibmcloud.ibm.com"}

	Context("create Service instance from ibmcloud.ibm.com WITHOUT external dependencies", func() {
		It("should correctly create the Service instance", func() {

			comp := test.LoadCompasable(dataDir + "comp.yaml")
			test.PostInNs(testContext, &comp, false, 0)
			Eventually(test.GetObject(testContext, &comp)).ShouldNot(BeNil())

			objNamespacedname := types.NamespacedName{Namespace: testContext.Namespace(), Name: "mymessagehub"}
			unstrObj := unstructured.Unstructured{}
			unstrObj.SetGroupVersionKind(groupVersionKind)
			klog.V(5).Infof("Get Object %s\n", objNamespacedname)
			Eventually(test.GetUnstructuredObject(testContext, objNamespacedname, &unstrObj)).Should(Succeed())
			Eventually(test.GetState(testContext, &comp)).Should(Equal(OnlineStatus))

		})

		It("should delete the Composable and Service instances", func() {
			By("Delete the Composable object")
			comp := test.LoadCompasable(dataDir + "comp.yaml")
			test.DeleteInNs(testContext, &comp, false)
			Eventually(test.GetObject(testContext, &comp)).Should(BeNil())

			// TODO update for external test only
			/*
				By("Validate that the underlying object is deleted too")
				objNamespacedname := types.NamespacedName{Namespace: testContext.Namespace(), Name: "mymessagehub"}
				unstrObj := unstructured.Unstructured{}
				unstrObj.SetGroupVersionKind(groupVersionKind)
				Eventually(func() bool {
					err := test.GetUnstructuredObject(testContext, objNamespacedname, &unstrObj)()
					return errors.IsNotFound(err)
				}).Should(BeTrue())
			*/
		})

	})

	Context("create Service instance from ibmcloud.ibm.com WITH external dependencies", func() {
		var objNamespacedname types.NamespacedName

		BeforeEach(func() {
			obj := test.LoadObject(dataDir+"mysecret.yaml", &v1.Secret{})
			test.PostInNs(testContext, obj, false, 0)
			objNamespacedname = types.NamespacedName{Namespace: testContext.Namespace(), Name: "mymessagehub"}
			Eventually(test.GetObject(testContext, obj)).ShouldNot(BeNil())
		})

		AfterEach(func() {
			obj := test.LoadObject(dataDir+"mysecret.yaml", &v1.Secret{})
			test.DeleteInNs(testContext, obj, false)
		})

		It("should correctly create the Service instance according to parameters from a Secret object", func() {
			By("deploy Composable comp1.yaml")
			comp := test.LoadCompasable(dataDir + "comp1.yaml")
			test.PostInNs(testContext, &comp, false, 0)
			Eventually(test.GetObject(testContext, &comp)).ShouldNot(BeNil())

			By("get underlying object - Service.ibmcloud.ibm.com/v1alpha1")
			unstrObj := unstructured.Unstructured{}
			unstrObj.SetGroupVersionKind(groupVersionKind)
			klog.V(5).Infof("Get Object %s\n", objNamespacedname)
			Eventually(test.GetUnstructuredObject(testContext, objNamespacedname, &unstrObj)).Should(Succeed())

			By("validate service plan")
			Ω(getPlan(unstrObj.Object)).Should(Equal("standard"))

			By("Reload the Composable object")
			Eventually(test.GetObject(testContext, &comp)).ShouldNot(BeNil())

			By("validate that Composable object status is Online")
			Eventually(test.GetState(testContext, &comp)).Should(Equal(OnlineStatus))

			By("delete the composable object")
			test.DeleteInNs(testContext, &comp, false)
			Eventually(test.GetObject(testContext, &comp)).Should(BeNil())
		})

		It("should correctly create the Service instance according to parameters from a ConfigMap", func() {
			By("Deploy the myconfigmap  ConfigMap")
			obj := test.LoadObject(dataDir+"myconfigmap.yaml", &v1.ConfigMap{})
			test.PostInNs(testContext, obj, false, 0)
			Eventually(test.GetObject(testContext, obj)).ShouldNot(BeNil())

			By("deploy Composable comp2.yaml ")
			comp := test.LoadCompasable(dataDir + "comp2.yaml")
			test.PostInNs(testContext, &comp, false, 0)
			Eventually(test.GetObject(testContext, &comp)).ShouldNot(BeNil())

			By("get underlying object - Service.ibmcloud.ibm.com/v1alpha1")
			unstrObj := unstructured.Unstructured{}
			unstrObj.SetGroupVersionKind(groupVersionKind)
			klog.V(5).Infof("Get Object %s\n", objNamespacedname)
			Eventually(test.GetUnstructuredObject(testContext, objNamespacedname, &unstrObj)).Should(Succeed())

			By("validate service plan")
			Ω(getPlan(unstrObj.Object)).Should(Equal("standard"))

			By("Reload the Composable object")
			Eventually(test.GetObject(testContext, &comp)).ShouldNot(BeNil())

			By("validate that Composable object status is Online")
			Eventually(test.GetState(testContext, &comp)).Should(Equal(OnlineStatus))

			By("delete the composable object")
			test.DeleteInNs(testContext, &comp, false)
			Eventually(test.GetObject(testContext, &comp)).Should(BeNil())
		})
	})

})

// returns service plan of Service.ibmcloud.ibm.com
func getPlan(objMap map[string]interface{}) string {
	if spec, ok := objMap[spec].(map[string]interface{}); ok {
		if plan, ok := spec["plan"].(string); ok {
			return plan
		}
	}
	return ""
}