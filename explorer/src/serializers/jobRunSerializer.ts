import {
  Serializer as JSONAPISerializer,
  SerializerOptions
} from 'jsonapi-serializer'
import { JobRun } from '../entity/JobRun'

export const BASE_ATTRIBUTES = [
  'chainlinkNode',
  'runId',
  'jobId',
  'status',
  'type',
  'requestId',
  'txHash',
  'requester',
  'error',
  'createdAt',
  'finishedAt'
]

export const chainlinkNode = {
  ref: 'id',
  attributes: ['name']
}

export const taskRuns = {
  ref: 'id',
  attributes: [
    'jobRunId',
    'index',
    'type',
    'status',
    'error',
    'transactionHash',
    'transactionStatus'
  ]
}

const ETHERSCAN_HOST = process.env.ETHERSCAN_HOST || 'ropsten.etherscan.io'

const jobRunSerializer = (run: JobRun) => {
  const opts = {
    attributes: BASE_ATTRIBUTES.concat(['taskRuns']),
    keyForAttribute: 'camelCase',
    chainlinkNode: chainlinkNode,
    taskRuns: taskRuns,
    meta: {
      etherscanHost: ETHERSCAN_HOST
    }
  } as SerializerOptions

  return new JSONAPISerializer('job_runs', opts).serialize(run)
}

export default jobRunSerializer
