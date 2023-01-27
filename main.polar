allow(actor, action, resource) if
  has_permission(actor, action, resource);

actor User {}

# Teams
resource Team {
    permissions = ["read", "edit", "admin"];
    roles = ["guest", "researcher", "maintainer", "owner"];

    relations = { parent: Team };

    "read" if "guest";
    "edit" if "maintainer";
    "admin" if "owner";

    "guest" if "researcher";
    "researcher" if "maintainer";
    "maintainer" if "owner";

    # Check hierarchy
    "guest" if "guest" on "parent";
    "researcher" if "researcher" on "parent";
    "maintainer" if "maintainer" on "parent";
    "owner" if "owner" on "parent";
}

has_relation(parent: Team, "parent", team: Team) if
    team.Parent = parent;

has_role(user: User, name: String, team: Team) if
    role in user.Teams and
    role.Role = name and
    role.TeamID = team.ID;
