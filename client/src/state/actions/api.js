import ActionsTypes from '../../constants/ActionsTypes';

const api = {
  pingApi: () => ({
    type: ActionsTypes.PING_API,
    payload: {},
  }),
  pingApiFailure: (error) => ({
    type: ActionsTypes.PING_API_FAILURE,
    payload: { error },
  }),

  listAgents: () => ({
    type: ActionsTypes.LIST_AGENTS,
    payload: {},
  }),
  listAgentsSuccess: (request, response) => ({
    type: ActionsTypes.LIST_AGENTS_SUCCESS,
    payload: { request, response },
  }),

  getAgent: (org, cn) => ({
    type: ActionsTypes.GET_AGENT,
    payload: { org, cn },
  }),
  getAgentSuccess: (request, response) => ({
    type: ActionsTypes.GET_AGENT_SUCCESS,
    payload: { request, response },
  }),
};

export default api;
