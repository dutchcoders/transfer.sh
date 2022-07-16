import { Router } from "itty-router";

import { IRequest, IMethods } from "./types";
import api from "./api";

const router = Router<IRequest, IMethods>();

router.get("/get/:id/:file", api.getFile);
router.get("/:id/:file", api.getFileOrPassthrough);
router.all("*", api.passthrough);

export default {
  fetch: router.handle,
};
