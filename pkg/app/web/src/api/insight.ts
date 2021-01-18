import { apiClient, apiRequest } from "./client";
import {
  GetInsightMetricsDataRequest,
  GetInsightDataPointsResponse,
} from "pipe/pkg/app/web/api_client/service_pb";

export const getInsightData = ({
  applicationId,
  dataPointCount,
  metricsKind,
  rangeFrom,
  step,
}: GetInsightMetricsDataRequest.AsObject): Promise<
  GetInsightDataPointsResponse.AsObject
> => {
  const req = new GetInsightMetricsDataRequest();
  req.setApplicationId(applicationId);
  req.setDataPointCount(dataPointCount);
  req.setMetricsKind(metricsKind);
  req.setRangeFrom(rangeFrom);
  req.setStep(step);
  console.log(step);

  return apiRequest(req, apiClient.getInsightMetricsData);
};
