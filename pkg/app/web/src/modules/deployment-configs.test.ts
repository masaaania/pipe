import {
  deploymentConfigsSlice,
  DeploymentConfigsState,
  clearTemplateTarget,
  fetchTemplateList,
} from "./deployment-configs";
import { dummyDeploymentConfigTemplates } from "../__fixtures__/dummy-deployment-config";
import { addApplication } from "./applications";

const initialState: DeploymentConfigsState = {
  templates: {},
  targetApplicationId: null,
};

describe("deploymentConfigsSlice reducer", () => {
  it("should return the initial state", () => {
    expect(
      deploymentConfigsSlice.reducer(undefined, {
        type: "TEST_ACTION",
      })
    ).toEqual(initialState);
  });

  it(`should handle ${clearTemplateTarget.type}`, () => {
    expect(
      deploymentConfigsSlice.reducer(
        { ...initialState, targetApplicationId: "application-1" },
        {
          type: clearTemplateTarget.type,
        }
      )
    ).toEqual(initialState);
  });

  describe("fetchTemplateList", () => {
    it(`should handle ${fetchTemplateList.fulfilled.type}`, () => {
      expect(
        deploymentConfigsSlice.reducer(initialState, {
          type: fetchTemplateList.fulfilled.type,
          meta: {
            arg: {
              applicationId: "application-1",
            },
          },
          payload: dummyDeploymentConfigTemplates,
        })
      ).toEqual({
        ...initialState,
        templates: {
          "application-1": dummyDeploymentConfigTemplates,
        },
      });
    });
  });

  describe("addApplication", () => {
    it(`should handle ${addApplication.fulfilled.type}`, () => {
      expect(
        deploymentConfigsSlice.reducer(initialState, {
          type: addApplication.fulfilled.type,
          payload: "application-id",
        })
      ).toEqual({
        ...initialState,
        targetApplicationId: "application-id",
      });
    });
  });
});
