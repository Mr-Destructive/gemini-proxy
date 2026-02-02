import inspect
from typing import Dict, Any, Callable, List


def function_to_tool_schema(fn: Callable) -> Dict[str, Any]:
    """Convert a Python function to a tool schema."""
    sig = inspect.signature(fn)
    params = sig.parameters
    description = inspect.getdoc(fn) or ""

    def pytype_to_jsonschema(ann):
        if ann == int:
            return "integer"
        elif ann == float:
            return "number"
        elif ann == bool:
            return "boolean"
        elif ann == str:
            return "string"
        elif ann == list:
            return "array"
        elif ann == dict:
            return "object"
        return "string"

    input_schema = {
        "$schema": "http://json-schema.org/draft-07/schema#",
        "type": "object",
        "properties": {},
    }
    required = []
    
    for name, param in params.items():
        ann = param.annotation if param.annotation != inspect.Parameter.empty else str
        param_type = pytype_to_jsonschema(ann)
        if param.default == inspect.Parameter.empty:
            required.append(name)
        input_schema["properties"][name] = {
            "type": param_type,
            "description": "",
        }
    
    if required:
        input_schema["required"] = required
    
    return {
        "name": fn.__name__,
        "description": description,
        "input_schema": input_schema,
    }


class ToolRegistry:
    """Registry for managing tools/functions."""
    
    def __init__(self):
        self.tools: Dict[str, Callable] = {}
        self.schemas: Dict[str, Dict[str, Any]] = {}
    
    def register(self, fn: Callable) -> Callable:
        """Register a function as a tool."""
        name = fn.__name__
        self.tools[name] = fn
        self.schemas[name] = function_to_tool_schema(fn)
        return fn
    
    def get_tools(self) -> List[Dict[str, Any]]:
        """Get all tool schemas."""
        return list(self.schemas.values())
    
    def call(self, tool_name: str, **kwargs) -> Any:
        """Call a registered tool."""
        if tool_name not in self.tools:
            raise ValueError(f"Tool '{tool_name}' not found")
        return self.tools[tool_name](**kwargs)
