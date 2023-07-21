package graphvent

import (
  "github.com/graphql-go/graphql"
  "reflect"
)

var gql_interface_graph_node *graphql.Interface = nil
func GQLInterfaceNode() *graphql.Interface {
  if gql_interface_graph_node == nil {
    gql_interface_graph_node = graphql.NewInterface(graphql.InterfaceConfig{
      Name: "Node",
      ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
        ctx, ok := p.Context.Value("graph_context").(*Context)
        if ok == false {
          return nil
        }

        valid_nodes := ctx.GQL.ValidNodes
        node_type := ctx.GQL.NodeType
        p_type := reflect.TypeOf(p.Value)

        for key, value := range(valid_nodes) {
          if p_type == key {
            return value
          }
        }

        if p_type.Implements(node_type) {
          return GQLTypeGraphNode()
        }

        return nil
      },
      Fields: graphql.Fields{},
    })

    gql_interface_graph_node.AddFieldConfig("ID", &graphql.Field{
      Type: graphql.String,
    })
  }

  return gql_interface_graph_node
}

var gql_list_thread *graphql.List = nil
func GQLListThread() *graphql.List {
  if gql_list_thread == nil {
    gql_list_thread = graphql.NewList(GQLInterfaceThread())
  }
  return gql_list_thread
}

var gql_interface_thread *graphql.Interface = nil
func GQLInterfaceThread() *graphql.Interface {
  if gql_interface_thread == nil {
    gql_interface_thread = graphql.NewInterface(graphql.InterfaceConfig{
      Name: "Thread",
      ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
        ctx, ok := p.Context.Value("graph_context").(*Context)
        if ok == false {
          return nil
        }

        valid_threads := ctx.GQL.ValidThreads
        thread_type := ctx.GQL.ThreadType
        p_type := reflect.TypeOf(p.Value)

        for key, value := range(valid_threads) {
          if p_type == key {
            return value
          }
        }

        if p_type.Implements(thread_type) {
          return GQLTypeSimpleThread()
        }

        ctx.Log.Logf("gql", "Found no type that matches %+v: %+v", p_type, p_type.Implements(thread_type))
        return nil
      },
      Fields: graphql.Fields{},
    })

    gql_interface_thread.AddFieldConfig("ID", &graphql.Field{
      Type: graphql.String,
    })

    gql_interface_thread.AddFieldConfig("Name", &graphql.Field{
      Type: graphql.String,
    })

    gql_interface_thread.AddFieldConfig("State", &graphql.Field{
      Type: graphql.String,
    })

    gql_interface_thread.AddFieldConfig("Children", &graphql.Field{
      Type: GQLListThread(),
    })

    gql_interface_thread.AddFieldConfig("Parent", &graphql.Field{
      Type: GQLInterfaceThread(),
    })

    gql_interface_thread.AddFieldConfig("Requirements", &graphql.Field{
      Type: GQLListLockable(),
    })

    gql_interface_thread.AddFieldConfig("Dependencies", &graphql.Field{
      Type: GQLListLockable(),
    })

    gql_interface_thread.AddFieldConfig("Owner", &graphql.Field{
      Type: GQLInterfaceLockable(),
    })
  }

  return gql_interface_thread
}

var gql_list_lockable *graphql.List = nil
func GQLListLockable() *graphql.List {
  if gql_list_lockable == nil {
    gql_list_lockable = graphql.NewList(GQLInterfaceLockable())
  }
  return gql_list_lockable
}

var gql_interface_lockable *graphql.Interface = nil
func GQLInterfaceLockable() *graphql.Interface {
  if gql_interface_lockable == nil {
    gql_interface_lockable = graphql.NewInterface(graphql.InterfaceConfig{
      Name: "Lockable",
      ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
        ctx, ok := p.Context.Value("graph_context").(*Context)
        if ok == false {
          return nil
        }

        valid_lockables := ctx.GQL.ValidLockables
        lockable_type := ctx.GQL.LockableType
        p_type := reflect.TypeOf(p.Value)

        for key, value := range(valid_lockables) {
          if p_type == key {
            return value
          }
        }

        if p_type.Implements(lockable_type) {
          return GQLTypeSimpleLockable()
        }
        return nil
      },
      Fields: graphql.Fields{},
    })

    gql_interface_lockable.AddFieldConfig("ID", &graphql.Field{
      Type: graphql.String,
    })

    gql_interface_lockable.AddFieldConfig("Name", &graphql.Field{
      Type: graphql.String,
    })

    if gql_list_lockable == nil {
      gql_list_lockable = graphql.NewList(gql_interface_lockable)
    }

    gql_interface_lockable.AddFieldConfig("Requirements", &graphql.Field{
      Type: gql_list_lockable,
    })

    gql_interface_lockable.AddFieldConfig("Dependencies", &graphql.Field{
      Type: gql_list_lockable,
    })

    gql_interface_lockable.AddFieldConfig("Owner", &graphql.Field{
      Type: gql_interface_lockable,
    })

  }

  return gql_interface_lockable
}

var gql_list_user *graphql.List = nil
func GQLListUser() *graphql.List {
  if gql_list_user == nil {
    gql_list_user = graphql.NewList(GQLTypeUser())
  }
  return gql_list_user
}

var gql_type_user *graphql.Object = nil
func GQLTypeUser() * graphql.Object {
  if gql_type_user == nil {
    gql_type_user = graphql.NewObject(graphql.ObjectConfig{
      Name: "User",
      Interfaces: []*graphql.Interface{
        GQLInterfaceNode(),
        GQLInterfaceLockable(),
      },
      IsTypeOf: func(p graphql.IsTypeOfParams) bool {
        ctx, ok := p.Context.Value("graph_context").(*Context)
        if ok == false {
          return false
        }

        lockable_type := ctx.GQL.LockableType
        value_type := reflect.TypeOf(p.Value)

        if value_type.Implements(lockable_type) {
          return true
        }

        return false
      },
      Fields: graphql.Fields{},
    })

    gql_type_user.AddFieldConfig("ID", &graphql.Field{
      Type: graphql.String,
      Resolve: GQLNodeID,
    })

    gql_type_user.AddFieldConfig("Name", &graphql.Field{
      Type: graphql.String,
      Resolve: GQLLockableName,
    })

    gql_type_user.AddFieldConfig("Requirements", &graphql.Field{
      Type: GQLListLockable(),
      Resolve: GQLLockableRequirements,
    })

    gql_type_user.AddFieldConfig("Owner", &graphql.Field{
      Type: GQLInterfaceLockable(),
      Resolve: GQLLockableOwner,
    })

    gql_type_user.AddFieldConfig("Dependencies", &graphql.Field{
      Type: GQLListLockable(),
      Resolve: GQLLockableDependencies,
    })
  }
  return gql_type_user
}

var gql_type_gql_thread *graphql.Object = nil
func GQLTypeGQLThread() * graphql.Object {
  if gql_type_gql_thread == nil {
    gql_type_gql_thread = graphql.NewObject(graphql.ObjectConfig{
      Name: "GQLThread",
      Interfaces: []*graphql.Interface{
        GQLInterfaceNode(),
        GQLInterfaceThread(),
        GQLInterfaceLockable(),
      },
      IsTypeOf: func(p graphql.IsTypeOfParams) bool {
        _, ok := p.Value.(*GQLThread)
        return ok
      },
      Fields: graphql.Fields{},
    })

    gql_type_gql_thread.AddFieldConfig("ID", &graphql.Field{
      Type: graphql.String,
      Resolve: GQLNodeID,
    })

    gql_type_gql_thread.AddFieldConfig("Name", &graphql.Field{
      Type: graphql.String,
      Resolve: GQLLockableName,
    })

    gql_type_gql_thread.AddFieldConfig("State", &graphql.Field{
      Type: graphql.String,
      Resolve: GQLThreadState,
    })

    gql_type_gql_thread.AddFieldConfig("Children", &graphql.Field{
      Type: GQLListThread(),
      Resolve: GQLThreadChildren,
    })

    gql_type_gql_thread.AddFieldConfig("Parent", &graphql.Field{
      Type: GQLInterfaceThread(),
      Resolve: GQLThreadParent,
    })

    gql_type_gql_thread.AddFieldConfig("Listen", &graphql.Field{
      Type: graphql.String,
      Resolve: GQLThreadListen,
    })

    gql_type_gql_thread.AddFieldConfig("Requirements", &graphql.Field{
      Type: GQLListLockable(),
      Resolve: GQLLockableRequirements,
    })

    gql_type_gql_thread.AddFieldConfig("Owner", &graphql.Field{
      Type: GQLInterfaceLockable(),
      Resolve: GQLLockableOwner,
    })

    gql_type_gql_thread.AddFieldConfig("Dependencies", &graphql.Field{
      Type: GQLListLockable(),
      Resolve: GQLLockableDependencies,
    })

    gql_type_gql_thread.AddFieldConfig("Users", &graphql.Field{
      Type: GQLListUser(),
      Resolve: GQLThreadUsers,
    })
  }
  return gql_type_gql_thread
}

var gql_type_simple_thread *graphql.Object = nil
func GQLTypeSimpleThread() * graphql.Object {
  if gql_type_simple_thread == nil {
    gql_type_simple_thread = graphql.NewObject(graphql.ObjectConfig{
      Name: "SimpleThread",
      Interfaces: []*graphql.Interface{
        GQLInterfaceNode(),
        GQLInterfaceThread(),
        GQLInterfaceLockable(),
      },
      IsTypeOf: func(p graphql.IsTypeOfParams) bool {
        ctx, ok := p.Context.Value("graph_context").(*Context)
        if ok == false {
          return false
        }

        thread_type := ctx.GQL.ThreadType

        value_type := reflect.TypeOf(p.Value)

        if value_type.Implements(thread_type) {
          return true
        }

        return false
      },
      Fields: graphql.Fields{},
    })
    gql_type_simple_thread.AddFieldConfig("ID", &graphql.Field{
      Type: graphql.String,
      Resolve: GQLNodeID,
    })

    gql_type_simple_thread.AddFieldConfig("Name", &graphql.Field{
      Type: graphql.String,
      Resolve: GQLLockableName,
    })

    gql_type_simple_thread.AddFieldConfig("State", &graphql.Field{
      Type: graphql.String,
      Resolve: GQLThreadState,
    })

    gql_type_simple_thread.AddFieldConfig("Children", &graphql.Field{
      Type: GQLListThread(),
      Resolve: GQLThreadChildren,
    })

    gql_type_simple_thread.AddFieldConfig("Parent", &graphql.Field{
      Type: GQLInterfaceThread(),
      Resolve: GQLThreadParent,
    })

    gql_type_simple_thread.AddFieldConfig("Requirements", &graphql.Field{
      Type: GQLListLockable(),
      Resolve: GQLLockableRequirements,
    })

    gql_type_simple_thread.AddFieldConfig("Owner", &graphql.Field{
      Type: GQLInterfaceLockable(),
      Resolve: GQLLockableOwner,
    })

    gql_type_simple_thread.AddFieldConfig("Dependencies", &graphql.Field{
      Type: GQLListLockable(),
      Resolve: GQLLockableDependencies,
    })
  }
  return gql_type_simple_thread
}

var gql_type_simple_lockable *graphql.Object = nil
func GQLTypeSimpleLockable() * graphql.Object {
  if gql_type_simple_lockable == nil {
    gql_type_simple_lockable = graphql.NewObject(graphql.ObjectConfig{
      Name: "SimpleLockable",
      Interfaces: []*graphql.Interface{
        GQLInterfaceNode(),
        GQLInterfaceLockable(),
      },
      IsTypeOf: func(p graphql.IsTypeOfParams) bool {
        ctx, ok := p.Context.Value("graph_context").(*Context)
        if ok == false {
          return false
        }

        lockable_type := ctx.GQL.LockableType
        value_type := reflect.TypeOf(p.Value)

        if value_type.Implements(lockable_type) {
          return true
        }

        return false
      },
      Fields: graphql.Fields{},
    })

    gql_type_simple_lockable.AddFieldConfig("ID", &graphql.Field{
      Type: graphql.String,
      Resolve: GQLNodeID,
    })

    gql_type_simple_lockable.AddFieldConfig("Name", &graphql.Field{
      Type: graphql.String,
      Resolve: GQLLockableName,
    })

    gql_type_simple_lockable.AddFieldConfig("Requirements", &graphql.Field{
      Type: GQLListLockable(),
      Resolve: GQLLockableRequirements,
    })

    gql_type_simple_lockable.AddFieldConfig("Owner", &graphql.Field{
      Type: GQLInterfaceLockable(),
      Resolve: GQLLockableOwner,
    })

    gql_type_simple_lockable.AddFieldConfig("Dependencies", &graphql.Field{
      Type: GQLListLockable(),
      Resolve: GQLLockableDependencies,
    })
  }
  return gql_type_simple_lockable
}

var gql_type_simple_node *graphql.Object = nil
func GQLTypeGraphNode() * graphql.Object {
  if gql_type_simple_node == nil {
    gql_type_simple_node = graphql.NewObject(graphql.ObjectConfig{
      Name: "GraphNode",
      Interfaces: []*graphql.Interface{
        GQLInterfaceNode(),
      },
      IsTypeOf: func(p graphql.IsTypeOfParams) bool {
        ctx, ok := p.Context.Value("graph_context").(*Context)
        if ok == false {
          return false
        }

        node_type := ctx.GQL.NodeType
        value_type := reflect.TypeOf(p.Value)

        if value_type.Implements(node_type) {
          return true
        }

        return false
      },
      Fields: graphql.Fields{},
    })

    gql_type_simple_node.AddFieldConfig("ID", &graphql.Field{
      Type: graphql.String,
      Resolve: GQLNodeID,
    })

    gql_type_simple_node.AddFieldConfig("Name", &graphql.Field{
      Type: graphql.String,
      Resolve: GQLLockableName,
    })
  }

  return gql_type_simple_node
}

var gql_type_signal *graphql.Object = nil
func GQLTypeSignal() *graphql.Object {
  if gql_type_signal == nil {
    gql_type_signal = graphql.NewObject(graphql.ObjectConfig{
      Name: "SignalOut",
      IsTypeOf: func(p graphql.IsTypeOfParams) bool {
        _, ok := p.Value.(GraphSignal)
        return ok
      },
      Fields: graphql.Fields{},
    })

    gql_type_signal.AddFieldConfig("Type", &graphql.Field{
      Type: graphql.String,
      Resolve: GQLSignalType,
    })
    gql_type_signal.AddFieldConfig("Source", &graphql.Field{
      Type: graphql.String,
      Resolve: GQLSignalSource,
    })
    gql_type_signal.AddFieldConfig("Direction", &graphql.Field{
      Type: graphql.String,
      Resolve: GQLSignalDirection,
    })
    gql_type_signal.AddFieldConfig("String", &graphql.Field{
      Type: graphql.String,
      Resolve: GQLSignalString,
    })
  }
  return gql_type_signal
}

var gql_type_signal_input *graphql.InputObject = nil
func GQLTypeSignalInput() *graphql.InputObject {
  if gql_type_signal_input == nil {
    gql_type_signal_input = graphql.NewInputObject(graphql.InputObjectConfig{
      Name: "SignalIn",
      Fields: graphql.InputObjectConfigFieldMap{},
    })
    gql_type_signal_input.AddFieldConfig("Type", &graphql.InputObjectFieldConfig{
      Type: graphql.String,
      DefaultValue: "cancel",
    })
    gql_type_signal_input.AddFieldConfig("Direction", &graphql.InputObjectFieldConfig{
      Type: graphql.String,
      DefaultValue: "down",
    })
  }
  return gql_type_signal_input
}

